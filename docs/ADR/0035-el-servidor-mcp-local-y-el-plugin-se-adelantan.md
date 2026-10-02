# 0035 · El servidor MCP local y el plugin se adelantan: kitlegal donde no hay terminal, y todo en el equipo

- **Estado**: aceptada
- **Fecha**: 2026-10-01
- **Hito**: transversal (tras v0.3.2, antes de H21). Saca del backlog de distribución el servidor MCP y el plugin
  y los numera H21 y H22, por delante de H20. Enmienda el ADR 0019 (el plugin y el MCP dejan de estar fuera de
  alcance), la última consecuencia del ADR 0034 y el principio VIII de la constitución (2.10.0): la tabla de
  comandos de una skill nombra cada operación como orden y como herramienta.

## Contexto y problema

El ADR 0034 dejó escrito que ninguna de las dos audiencias de kitlegal —la ciudadanía y los despachos— trabaja en
una terminal, y la bitácora (`docs/USO.md`, 2026-10-01) lo apuntó como la mayor distancia entre la web y quien la
lee. Los despachos buscan «claude cowork abogados», no Claude Code. El 2026-09-28 Jorge preguntó cómo usar kitlegal
desde Claude Cowork, y la respuesta fue que hoy no se puede:

- Cowork y el chat de Claude leen las skills y los plugins de la cuenta, no los del disco: lo que instala
  `kitlegal skills install` no les llega (ADR 0025).
- Aunque la skill llegara, su tabla de comandos dice «ejecuta `kitlegal boe articulo …`», y el shell de Cowork es
  una máquina virtual Linux donde el binario instalado en el equipo no está.

Lo que sí cruza esa frontera es un servidor MCP local: el programa corre en el equipo de la persona y el agente
recibe sus operaciones como herramientas. Estaba en el backlog de distribución desde el ADR 0013, con el plugin,
«para cuando alguien más tenga que usar esto o quieras usarlo fuera de Claude Code». Ese momento es este.

Lo que dice la documentación de cada programa, leída el 2026-10-01:

| Programa | Qué admite | Fuente |
| --- | --- | --- |
| Claude Cowork | Un plugin instalado en la cuenta carga sus skills. Un servidor MCP local del plugin —una orden que la app arranca, también un paquete `.mcpb`— carga «cuando la sesión de Cowork corre en tu equipo». Un `bin/` en la raíz del plugin impide instalarlo. | claude.com/docs/plugins/platform-support |
| Chat de Claude (web, móvil y escritorio) | Carga las skills del plugin e ignora su servidor MCP local. | claude.com/docs/plugins/platform-support |
| Claude Code | Carga todo. `mcpServers` del plugin admite un `.mcpb` por ruta dentro del plugin o por URL. | code.claude.com/docs/en/plugins/components |
| Instalar un plugin en Claude | *Customize > Plugins > Add*: subir un zip del plugin o añadir un marketplace por la dirección de su repositorio. El directorio de Anthropic, en planes de pago. | claude.com/docs/plugins/build |
| `.mcpb` (MCP Bundle) | Un zip con `manifest.json` (versión 0.3) y, con `server.type: binary`, el ejecutable dentro; `platform_overrides` distingue el sistema (`darwin`, `win32`, `linux`), no la arquitectura. No lleva skills. | github.com/modelcontextprotocol/mcpb, `MANIFEST.md` |
| App de escritorio de ChatGPT (Codex) | Servidores MCP locales por stdio en `~/.codex/config.toml` (`codex mcp add`, o *Settings > MCP servers*); la configuración la comparten la app, la CLI y la extensión de IDE. Lee el campo `instructions` del servidor. El sandbox no gobierna a los servidores MCP. | learn.chatgpt.com/docs/extend/mcp, …/sandboxing |
| ChatGPT en la web y en el móvil | Solo servidores MCP remotos. | learn.chatgpt.com/docs/extend/mcp |
| Antigravity | Servidores stdio en `~/.gemini/config/mcp_config.json` o `.agents/mcp_config.json`. | antigravity.google/docs/mcp |
| SDK de Go | `modelcontextprotocol/go-sdk` v1.8.0: servidor por stdio, herramientas con esquemas propios de entrada y de salida, `structuredContent`, `instructions`; atiende la especificación 2026-07-28 y las anteriores en el mismo proceso. Sin cgo. | github.com/modelcontextprotocol/go-sdk |
| goreleaser (edición libre) | No produce un `.mcpb` con `archives` (formatos cerrados, sin contenido por plantilla): hace falta un paso propio. Sí produce el binario universal de macOS y adjunta ficheros ya hechos a la release. Notariza en macOS sin un Mac (quill), con una cuenta de Apple Developer. | goreleaser.com |

La documentación no decía si macOS deja ejecutar el binario sin notarizar que llega dentro de un `.mcpb`, si
Cowork acepta el `.mcpb` del plugin por ruta igual que Claude Code, ni qué ve el servidor desde una sesión de
Cowork. Se comprobó a mano el mismo día, con un plugin de prueba hecho fuera del repositorio: una skill y un
`.mcpb` con un servidor mínimo en Go —binario universal de macOS, firmado solo por el enlazador, sin notarizar—
que pide al BOE el artículo 21 de la Ley 39/2015 y escribe un fichero en `~/.cache`. En la app de Claude
2.16120.0 para macOS:

- **El chat y Cowork son ya un solo modo**, y el otro es Code. Los plugins se instalan por modo: el zip subido
  estando en Code queda en `~/.claude/plugins/` y una conversación de chat no ve ni su skill ni su servidor.
- **Subido en el modo de chat, el plugin llega a la cuenta**: la app lo descarga de ahí, extrae el `.mcpb`,
  arranca el binario en el equipo y lo anuncia a la nube.
- **Sus herramientas solo llegan a una conversación que tiene elegida una carpeta del equipo** («Proyecto o
  carpeta»). Ahí la skill las llama, y el servidor leyó el BOE (200) y escribió y leyó en `~/.cache`. Una
  conversación sin carpeta corre en la nube con el equipo conectado como dispositivo, y recibe la skill y no las
  herramientas: así fue en cinco conversaciones nuevas, con dos plugins de nombre distinto y tras reiniciar la
  app, que en todas tenía el servidor en marcha y anunciado. Es lo que la documentación dice con «cuando la
  sesión corre en tu equipo». Hubo una excepción sin explicar: una conversación, al parecer abierta antes de
  instalar el plugin, las recibió sin carpeta, también las de un plugin instalado después.
- **El mismo `.mcpb`, instalado suelto con doble clic, sí llega a todas**: queda como extensión de escritorio,
  y una conversación nueva sin carpeta recibe sus herramientas y la skill del plugin las usa.
- **Sin herramientas, la regla aguanta**: la skill de prueba pedía una línea fija y no responder de memoria, y
  Claude la escribió sin buscar en la web ni usar el shell, que tenía a mano.
- **macOS no puso reparos**: el binario extraído no lleva el atributo de cuarentena, tampoco cuando el zip subido
  lo llevaba, puesto a mano como lo pondría un navegador.
- **Al instalar, la app avisa** en rojo de que el plugin «otorgará acceso a todo en tu computadora», y hay que
  aceptar.
- **Lo que el servidor encuentra**: su directorio de trabajo es `/`, también cuando la conversación tiene una
  carpeta, que no le llega; el ejecutable vive en un directorio de la app que cambia en cada arranque de esta; y
  la app negocia la versión 2025-11-25 del protocolo, no la vigente.
- **Una frase ambigua no activa la skill**: «prueba kitlegal» se entendió como «pasa los tests de kitlegal», por
  lo que la cuenta recuerda del proyecto. Invocada con `/`, respondió.

## Opciones consideradas

**Qué va después de H7.4.**

1. **H20, como estaba**, y el MCP en el backlog. Rechazada: lo que añadan H20, H8 y H9 no llega a ninguna de las
   dos audiencias mientras instalar pida una terminal y usarlo pida Claude Code. El roadmap se ordena por tiempo
   hasta el uso (ADR 0013).
2. **El servidor MCP y el plugin, y después H20.** Elegida.

**Dónde corre.**

1. **Un servidor remoto de kitlegal.** Es lo único que llega a ChatGPT y a Claude en la web y en el móvil.
   Aplazada por Jorge, que quiere pensarlo: kitlegal pasaría a recibir las preguntas de quien lo usa, lo que
   sustituye la promesa del ADR 0027 y lo convierte en responsable de ese tratamiento; y pagaría un servidor que
   otros consumen sin coste. No es una opción descartada: es otra decisión, con su ADR.
2. **Todo en el equipo de quien lo usa, con todas las funciones.** Elegida. Donde un servidor local no llega se
   dice que no es compatible y con qué sí funciona.

**Un hito o dos.**

1. **Uno, con el servidor, el paquete y el plugin.** Rechazada: son dos entregas que se miden distinto —la
   primera, con las evals en los dos modos; la segunda, con una instalación a mano—, y el servidor sirve por sí
   solo a quien ya tiene el binario.
2. **Dos**: H21, el servidor con las skills y las evals en los dos modos; H22, la instalación sin terminal.
   Elegida. La sección de H22 se detalla al cerrar H21.

**Cómo llega el binario a quien no abre una terminal** (H22).

1. **Un plugin con el binario en `bin/`.** Cowork y el chat no lo instalan.
2. **Un plugin cuyo servidor llama al `kitlegal` del `PATH`.** Exige haber instalado antes el binario, con la
   terminal. Sirve a quien ya lo tiene, no a la audiencia.
3. **El `.mcpb` suelto, con doble clic.** Lleva las herramientas y no las skills: el protocolo, que es el producto
   (principio VIII), no llega.
4. **Un plugin con las skills y el `.mcpb` dentro.** Un solo paso, pero la prueba a mano dejó ver que sus
   herramientas solo llegan a las conversaciones que tienen una carpeta: una pregunta suelta se queda sin ellas.
5. **Dos piezas de la misma etiqueta**: el `.mcpb` suelto, que da las herramientas a cualquier conversación, y un
   plugin solo con las skills. Elegida. El plugin no lleva el servidor, para que quien instale las dos no lo
   tenga dos veces en una conversación con carpeta.

**Cómo conviven la orden y la herramienta en una skill.**

1. **Una skill por modo.** Rechazada: dos protocolos que mantener iguales, y lo ganado de H7.2 a H7.4 medido solo
   en uno.
2. **Las skills dentro del servidor**, como `prompts` o `resources` MCP. Aplazada: ningún programa verificado los
   carga sin que la persona los elija.
3. **Una sola skill que nombra cada operación de las dos formas**, con la tabla generada desde `--describe`, y las
   evals medidas en los dos modos. Elegida.

## Decisión

- **H21 y H22 van por delante de H20**, con los dos números libres siguientes. El orden de la fase 1 pasa a ser
  H7.4 → H21 → H22 → H20 → H8 → H9.
- **H21: `kitlegal mcp serve`.** Un servidor MCP por stdio, único transporte, con una herramienta por cada verbo
  de consulta del registro, sus esquemas de `--describe` y el mismo sobre como resultado. Las dos skills nombran
  cada operación como herramienta y como orden, y las evals deciden en los dos modos.
- **H22: instalar sin terminal.** Cada release publica el `.mcpb`, que se instala con doble clic, y el plugin de
  Claude con las skills, que se instala desde *Customize > Plugins*. Son dos pasos y ninguno pide la terminal.
- **Todo corre en el equipo.** Ningún transporte de red, ningún servidor de kitlegal. La promesa del ADR 0027 no
  cambia.
- **Una skill sin herramienta ni orden lo dice.** Donde hay skill y no hay servidor —el plugin sin la extensión,
  la web y el móvil—, la respuesta dice que no ha podido consultar el BOE y cómo se arregla, con una forma fija,
  y no afirma nada de memoria.
- **El servidor no depende de dónde arranca**: ni del directorio de trabajo ni de la ruta de su ejecutable. La
  caché y el grafo siguen en `~/.cache/kitlegal/`.
- **Principio VIII, constitución 2.10.0**: la tabla de comandos nombra cada operación como orden
  (`kitlegal <applet> <verbo>`) y como herramienta (`<applet>_<verbo>`).

## Consecuencias

- H20, H8 y H9 se retrasan dos hitos. Nacen ya con su herramienta MCP, porque las herramientas salen del registro.
- La web y el README pueden decir con qué funciona: la app de escritorio de Claude y Claude Code; la app de
  escritorio de ChatGPT y Antigravity, con el binario instalado; y que ChatGPT y Claude en la web y en el móvil no
  son compatibles, porque solo admiten servidores remotos. Los textos de la web los cambia la persona (ADR 0024).
- El job de evals mide cada eval dos veces, una por modo, y los umbrales del ADR 0029 se cumplen en cada uno.
- Entra la dependencia `modelcontextprotocol/go-sdk`, ya prevista en el principio V para la distribución.
- Quien instale el plugin en su cuenta lo recibe también en Claude Code; si además tiene las skills instaladas con
  `kitlegal skills install`, las tiene dos veces. Lo documenta H22.
- Del backlog de distribución quedan los packs, `pkg/legalkit`, un plugin para la app de ChatGPT y el servidor
  remoto.

## Prueba con el binario de H21 (2026-10-02)

Con H21 en `main` (`097af64`), la prueba a mano de su aceptación en la app de escritorio de Claude para macOS, con
el binario universal de `main` dentro de un `.mcpb` hecho fuera del repositorio (manifiesto `0.3`, `server.type:
binary`, `${__dirname}/server/kitlegal mcp serve`; validado con la herramienta `mcpb`) e instalado con doble clic,
y la pregunta «¿qué dice el art. 21 de la Ley 39/2015?» en una conversación nueva sin carpeta:

- **Con la extensión sola**, la conversación llama a `boe_articulo` y devuelve el texto del BOE, pero sin la cita
  con su forma, con un enlace que no viene de ninguna herramienta y con vocabulario interno («respuesta en caché»).
  Las `instructions` del servidor, que piden esa forma con ese ejemplo, no bastan en esta app.
- **Con la extensión y un plugin solo con las skills**, subido como zip en el modo de chat, `boe-legislacion` se
  activa con la pregunta, llama a `boe_indice`, `boe_articulo` y `graph_check` y responde con
  `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]`.
- **Al instalar**, la app enseña una ficha con el aviso en rojo («otorgará a esta extensión acceso a todo lo que hay
  en tu computadora», desarrollador sin verificar por Anthropic), el icono genérico si el manifiesto no trae uno y
  la descripción cortada si es larga. macOS no puso reparos a un `.mcpb` creado en el mismo equipo.

Confirma la opción 5 —dos piezas de la misma etiqueta— con el binario real, y con ello se detalla H22
(`docs/ROADMAP.md`). Para el marketplace del plugin, la documentación de Claude Code (leída el 2026-10-02,
code.claude.com/docs/en/plugins/marketplace-reference y …/host-marketplace) da una fuente `archive` —un zip por
HTTPS con su `sha256`— y dice que quien lo añade recibe una versión nueva cuando cambia `version`.

La mitad de la aceptación de H21 en la app de escritorio de ChatGPT no se ha hecho, y Jorge decidió ese mismo día no
esperarla: lo que se da por soportado es lo que se puede probar, que hoy es la app de escritorio de Claude en macOS.
Lo demás —la extensión en Windows, la app de ChatGPT, Codex, Antigravity— se entrega o se documenta como «así se
haría, sin probar», pidiendo a quien lo intente que cuente si funciona, y pasa a soportado cuando alguien lo prueba.
El `.mcpb` de H22 lleva el binario de macOS y el de Windows, este sin probar. El marketplace del plugin vive en `jmorenobl/kitlegal-plugins`, un catálogo por agente que no
guarda ficheros de release.

## Pendiente de verificar (pasa a la aceptación humana de H22)

1. Un `.mcpb` y un zip descargados de verdad con un navegador, en un Mac distinto del que compiló el binario. Si
   macOS lo bloquea, hace falta notarizarlo: una cuenta de Apple Developer, que es una credencial y una decisión
   de la persona.
2. Cómo se actualizan una extensión instalada desde un fichero y un plugin subido como zip, y si un marketplace
   actualiza el plugin solo.
3. Qué responde Claude en la web y en el móvil, donde puede haber skill y no hay herramienta.
4. Lo mismo en Windows, donde nadie lo ha probado: el `.mcpb` lleva su binario para que alguien pueda.
5. Si la app de Claude admite añadir un marketplace cuya entrada es de fuente `archive`.
