---
id: H31
fase: "1"
depende_de: [H30]
---

# H31 · Las skills siguen la especificación Agent Skills: guía de autoría y gates

- **Objetivo**: skills cortas que guían y referencias que se cargan cuando hacen falta, como piden la especificación (agentskills.io) y la guía de autoría de Anthropic, en lugar de un `SKILL.md` que lo lleva todo.
- **Entrega**: una guía de autoría del proyecto sobre las dos fuentes; gates en `make ci` en lugar de «menos de 300 líneas»: la validación de la especificación (`skills-ref validate` o una comprobación equivalente sin red, según fije el plan), un presupuesto de tokens del cuerpo, la descripción en tercera persona y con margen bajo los 1024 caracteres, referencias a un solo nivel e índice en las de más de 100 líneas; la regla de `references/`: el protocolo se escribe a mano y lo que sale de `data/` sigue generado, con su cabecera; las tres skills reescritas con esa forma, una propuesta de cambio por skill.
- **Riesgos y dependencias**: la reescritura cambia lo que lee el modelo y la deciden las evals, sin regresión en activación ni en respuestas; depende de H30, sin el que cada skill reescrita mediría las tres.
