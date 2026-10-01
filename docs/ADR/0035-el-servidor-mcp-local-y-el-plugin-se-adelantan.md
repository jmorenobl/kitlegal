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

Sin comprobar, porque la documentación no lo dice: si macOS deja ejecutar el binario sin notarizar que llega
dentro de un `.mcpb`; si Cowork acepta el `.mcpb` del plugin por ruta igual que Claude Code; y qué ve el servidor
desde una sesión de Cowork (la caché del equipo, la red hacia el BOE).

## Opciones consideradas

**Qué va después de H7.4.**

1. **H20, como estaba**, y el MCP en el backlog. Rechazada: lo que añadan H20, H8 y H9 no llega a ninguna de las
   dos audiencias mientras instalar pida una terminal y usarlo pida Claude Code. El roadmap se ordena por tiempo
   hasta el uso (ADR 0013).
2. **El servidor MCP y el plugin, y después H20.** Elegida.

**Dónde corre.**

1. **Un servidor remoto de kitlegal.** Es lo único que llega a ChatGPT en la web y en el móvil y al chat de Claude.
   Aplazada por Jorge, que quiere pensarlo: kitlegal pasaría a recibir las preguntas de quien lo usa, lo que
   sustituye la promesa del ADR 0027 y lo convierte en responsable de ese tratamiento; y pagaría un servidor que
   otros consumen sin coste. No es una opción descartada: es otra decisión, con su ADR.
2. **Todo en el equipo de quien lo usa, con todas las funciones.** Elegida. Donde un servidor local no llega se
   dice que no es compatible y con qué sí funciona.

**Un hito o dos.**

1. **Uno, con el servidor, el paquete y el plugin.** Rechazada: el empaquetado depende de tres cosas que solo se
   comprueban a mano con el servidor ya hecho, y el servidor sirve por sí solo a quien ya tiene el binario.
2. **Dos**: H21, el servidor con las skills y las evals en los dos modos; H22, la instalación sin terminal.
   Elegida. La sección de H22 se detalla cuando esas comprobaciones estén hechas.

**Cómo llega el binario a quien no abre una terminal** (H22).

1. **Un plugin con el binario en `bin/`.** Cowork y el chat no lo instalan.
2. **Un plugin cuyo servidor llama al `kitlegal` del `PATH`.** Exige haber instalado antes el binario, con la
   terminal. Sirve a quien ya lo tiene, no a la audiencia.
3. **El `.mcpb` suelto, con doble clic.** Lleva las herramientas y no las skills: el protocolo, que es el producto
   (principio VIII), no llega.
4. **Un plugin con las skills y el `.mcpb` dentro**, generados de la misma etiqueta. Elegida, pendiente de las
   comprobaciones a mano.

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
- **H22: instalar sin terminal.** Cada release publica el `.mcpb` y el plugin de Claude con las skills y el
  servidor dentro, y quien no usa la terminal lo instala desde *Customize > Plugins*.
- **Todo corre en el equipo.** Ningún transporte de red, ningún servidor de kitlegal. La promesa del ADR 0027 no
  cambia.
- **Una skill sin herramienta ni orden lo dice.** En el chat de Claude las skills del plugin cargan y el servidor
  no: la respuesta dice que no ha podido consultar el BOE y cómo se instala, con una forma fija, y no afirma nada
  de memoria.
- **Principio VIII, constitución 2.10.0**: la tabla de comandos nombra cada operación como orden
  (`kitlegal <applet> <verbo>`) y como herramienta (`<applet>_<verbo>`).

## Consecuencias

- H20, H8 y H9 se retrasan dos hitos. Nacen ya con su herramienta MCP, porque las herramientas salen del registro.
- La web y el README pueden decir con qué funciona: Claude Cowork y Claude Code; la app de escritorio de ChatGPT y
  Antigravity, con el binario instalado; y que ChatGPT en la web y en el móvil y el chat de Claude no son
  compatibles, porque solo admiten servidores remotos. Los textos de la web los cambia la persona (ADR 0024).
- El job de evals mide cada eval dos veces, una por modo, y los umbrales del ADR 0029 se cumplen en cada uno.
- Entra la dependencia `modelcontextprotocol/go-sdk`, ya prevista en el principio V para la distribución.
- Quien instale el plugin en su cuenta lo recibe también en Claude Code; si además tiene las skills instaladas con
  `kitlegal skills install`, las tiene dos veces. Lo documenta H22.
- Del backlog de distribución quedan los packs, `pkg/legalkit`, un plugin para la app de ChatGPT y el servidor
  remoto.

## Pendiente de verificar (antes de detallar H22)

Con el binario de H21, a mano y en un Mac sin kitlegal instalado:

1. Que macOS ejecuta el binario que llega dentro del `.mcpb`. Si lo bloquea, hace falta notarizarlo: una cuenta
   de Apple Developer, que es una credencial y una decisión de la persona.
2. Que Cowork arranca el servidor de un plugin subido como zip con el `.mcpb` dentro.
3. Que, desde una sesión de Cowork, el servidor lee el BOE y escribe la caché y el grafo en el equipo.
4. Qué responde el chat de Claude con el plugin instalado, donde hay skill y no hay herramienta.
