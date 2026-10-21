# Distribución

Hojas: solo necesitan applets y `--describe`; entran cuando alguien más tenga que usar esto o quieras usarlo fuera de Claude Code.

- ~~`kitlegal mcp serve --stdio`~~ y ~~`plugin/`~~: son H21 y H22 desde el 2026-10-01 (ADR 0035).
- Mecanismo de packs (`packs/<vertical>/pack.yaml`, `legal` base) sobre el applet `skills` de H19 y el plugin de H22: `kitlegal skills install <pack>` instala un pack entero; `doctor` lanza además `verify-sources`; un plugin por pack, generado desde `packs/`. Winget cuando haya un usuario en Windows. (`--host antigravity` ya está, ADR 0025; Codex lee `.agents/skills/` y no lo necesita.) [enmienda 2026-10-21: entra en la fase 3 con la segunda vertical.]
- Un plugin para la app de escritorio de ChatGPT (Codex), con las skills y el servidor de H21: admite servidores stdio dentro de un plugin y lee los marketplaces de Claude, pero su directorio público no acepta servidores locales ni `.mcpb` (ADR 0035). Condición de entrada: comprobar a mano que un plugin con el binario dentro se instala y arranca sin terminal.
- Que el servidor de H21 sirva las skills empotradas como `prompts` o `resources` MCP, para los agentes que solo tienen el servidor. Condición de entrada: un agente verificado que los cargue sin que la persona los elija.- `pkg/legalkit`: facade mínima (`Resolve(cita)`, `Articulo(id, bloque)`, tipos de dominio), compromiso de compatibilidad documentado, `apidiff` en CI, ejemplo compilable. Release `v1.0.0`.
