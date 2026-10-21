---
id: H38
fase: "1"
---

# H38 · Las skills instaladas se actualizan con el binario

- **Objetivo**: que actualizar el binario con el gestor de paquetes no deje las skills globales en la versión anterior. El 2026-10-21, `brew upgrade` dejó el binario en v0.8.0 y `~/.agents/skills` en v0.7.0; el aviso de H19 lo ve el agente, no la persona, y no dice que la conversación ya cargó las skills anteriores.
- **Entrega**: en el ámbito global, si los ficheros instalados coinciden con las huellas de su manifiesto, el binario reinstala lo empotrado sin preguntar y lo dice en una línea («se aplican en la próxima conversación»); si no coinciden o no puede escribir —dentro de un sandbox—, avisa como hoy; en el ámbito local, solo avisa; se comprueba una vez por sesión. Cambia el contrato del aviso de H19 («nunca cambia la invocación»): ADR propio. Comprobar, antes de fijarlo, si un binario del plugin compara con un `~/.agents/skills` que la app no lee.
- **Controles**: e2e de las cuatro salidas (reinstala, ficheros cambiados, sin permiso de escritura, ámbito local) y de que `--dry-run` no escribe.
