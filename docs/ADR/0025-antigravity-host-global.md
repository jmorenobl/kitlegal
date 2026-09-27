# 0025 · Antigravity como host global: qué lee cada agente

- **Estado**: aceptada
- **Fecha**: 2026-09-27
- **Hito**: transversal (tras v0.2.0). Resuelve el «Pendiente de verificar» del ADR 0019 sobre Codex y Antigravity y
  amplía su contrato de hosts (FR-020 a FR-023 de H19): `--host` admite `antigravity` y se puede repetir.

## Contexto y problema

El ADR 0019 instala las skills en un directorio neutro por ámbito, `.agents/skills/` o `~/.agents/skills/`, y trata
como host, con un enlace relativo, a cada agente que no lo lee. Dejó pendiente comprobar si Codex y Antigravity lo
leen, con esta regla: si no, «ganan un valor de `--host`, no una copia». Comprobado el 2026-09-27:

- **Codex** lee `.agents/skills/` en cada directorio, desde el de trabajo hasta la raíz del repositorio, y
  `~/.agents/skills/` en global (documentación de Codex, *Build skills*; `~/.codex/skills/` sigue cargándose, marcado
  como obsoleto en su código). No necesita host.
- **Antigravity** —la app de escritorio 2.0, el IDE y el CLI `agy`— lee `<workspace>/.agents/skills/` en el proyecto,
  pero en global lee `~/.gemini/config/skills/` y no `~/.agents/skills/` (documentación de Antigravity, *Skills*; su
  CLI documenta además `~/.gemini/antigravity-cli/skills/`). Comprobado con `agy` 1.2.10: carga las skills de
  `~/.gemini/config/skills/`, también las que son enlaces relativos a `~/.agents/skills/`, y no carga las que solo
  están en `~/.agents/skills/`.
- **Claude Code** lee `.claude/skills/` y `~/.claude/skills/`, como ya estaba comprobado. **Cowork** y el chat de
  Claude cargan las skills de la cuenta (*Customize*), no las del disco: `kitlegal skills install` no les llega.

Así, `kitlegal skills install -g` dejaba las skills donde Antigravity no las ve.

## Opciones consideradas

1. **Documentar un enlace a mano.** Rechazada: es lo que el applet existe para no pedir, y `doctor` no lo vería.
2. **Copiar las skills en `~/.gemini/config/skills/`.** Rechazada por el ADR 0019: los hosts son enlaces, y la copia
   solo es el recurso cuando el sistema no deja enlazar.
3. **Un host `antigravity`, con su directorio de skills solo en global, sobre un registro de hosts** que sustituye al
   host `claude` fijo. Elegida.

## Decisión

1. Un **host** es un agente que no lee el directorio neutro: su nombre, su directorio de configuración —cuya
   existencia decide si se enlaza sin `--host`—, su directorio de skills y en qué ámbitos lo tiene.
   - `claude`: `.claude` y `.claude/skills`, en local y en global.
   - `antigravity`: `.gemini/config` y `.gemini/config/skills`, solo en global. En local lee el directorio neutro.
   - Los que leen el directorio neutro en los dos ámbitos, como Codex, no son hosts.
2. **Sin `--host`**, se enlaza en un host si su directorio de configuración y los de encima son directorios reales:
   un `~/.gemini/` sin `config/`, el que deja Gemini CLI, no cuenta. **`--host`** admite `claude` y `antigravity`, se
   puede repetir con un valor por bandera, sin partirlo por comas, y fuerza el host aunque falte su cadena, que se
   crea. `--host antigravity` en local se admite y no enlaza nada.
3. El **enlace** es relativo, con un `..` por cada elemento del directorio de skills: `../../.agents/skills/<skill>`
   desde `.claude/skills` y `../../../.agents/skills/<skill>` desde `.gemini/config/skills`.
4. Cada directorio de la raíz al de skills del host tiene que ser un directorio real o no existir, como `.claude` hasta
   ahora (FR-027, FR-028): un `~/.gemini` que es un enlace de un repositorio de dotfiles no se atraviesa; sin
   `--host` cuenta como ausente, y con `--host antigravity` es un conflicto.
5. El **manifiesto** admite en `hosts` la clave `antigravity`, con la ruta `.gemini/config/skills/<skill>`. Un
   manifiesto que declara una entrada de un host que el ámbito no tiene —cualquiera con `--dir`, `antigravity` en
   local— hace el ámbito ilegible, como ya lo hacía con `--dir`.
6. En la **salida**, `host` es `claude` o `antigravity`, y las entradas de cada skill van en el orden de los hosts,
   `claude` primero. La orden que da `doctor` lleva un `--host` por cada host en el que alguna de sus skills tiene una
   entrada declarada, para reinstalar aunque falte su directorio.

## Consecuencias

- `kitlegal skills install -g` deja las skills a la vista de Antigravity cuando existe `~/.gemini/config/`; también
  el `make install` de quien desarrolla.
- Contrato: `--describe` de `skills install` declara `host` como una lista, y el `host` de la salida admite
  `antigravity`. El mensaje del rechazo cambia a «los hosts admitidos son claude y antigravity».
- Un binario anterior no lee un manifiesto global que declara `antigravity`: lo da por ilegible (código 7) y no cambia
  nada. Volver a una versión anterior pide retirar esas entradas a mano (`rm -r ~/.gemini/config/skills/<skill>` y
  quitarlas de `~/.agents/skills/kitlegal.json`) o reinstalar desde cero. Solo afecta al ámbito global.
- Codex ejecuta las órdenes en un entorno aislado sin red: pide permiso en cada consulta de `kitlegal` al BOE salvo
  con una regla que lo permita. Lo explica el README; no es cosa del instalador tocar la configuración de Codex.
- Cowork y el chat de Claude siguen fuera: su vía es un plugin con las skills y un servidor MCP local (backlog ·
  distribución), no un host.
