# 0020 · Fusión y release: los decide la persona, los ejecuta el agente

- **Estado**: aceptada
- **Fecha**: 2026-09-27
- **Hito**: transversal (constitución 2.2.0, tras H19). Matiza el ADR 0007 y el ADR 0018 en quién ejecuta la fusión
  y el release; no cambia quién los decide.

## Contexto y problema

La constitución (principio 5 de «Flujo de trabajo y gates») dice que la fusión a `main` y el release «son siempre
acciones humanas», y `.claude/settings.json` lo hacía cumplir denegando `gh pr merge` y `curl` a todo agente. El
propósito era que el run desatendido no pudiera fusionar ni publicar por su cuenta. Pero la denegación alcanzaba
también a la sesión interactiva: al cerrar H19, la persona tuvo que ejecutar a mano la fusión, el cambio de
visibilidad del repositorio, la protección de `main`, la etiqueta `v0.1.0` y la prueba de `curl … | sh`, copiando
órdenes que no tenía por qué conocer. El agente era quien sabía qué ejecutar y la persona hacía de manos.

Lo que tiene que ser humano es la **decisión** (leer el informe final y decir «fusiona», «publica»), no teclear la
orden.

## Opciones consideradas

1. **Mantener la denegación.** Correcto para el run, pero obliga a la persona a ejecutar órdenes que decide otro.
2. **Permitir esas órdenes sin más.** El run desatendido podría fusionar o etiquetar: rompe el ADR 0018.
3. **Pasarlas de `deny` a `ask`.** En una sesión interactiva, Claude Code pide confirmación para cada una: la persona
   la da en ese momento, con la orden exacta delante. En una sesión headless (`claude -p`, la del workflow `hito`)
   no hay nadie que confirme y `ask` equivale a denegar, así que el run sigue sin poder hacerlo.

## Decisión

Opción 3.

- La fusión a `main` y el release los **decide** siempre una persona. El agente los **ejecuta** solo a petición
  expresa de la persona en una sesión interactiva, con confirmación de cada orden, y nunca dentro de un run.
- En `.claude/settings.json` pasan a `ask`: `gh pr merge`, `gh repo edit`, `gh api`, el push de una etiqueta de
  release con `KITLEGAL_PUSH_HUMANO=1` y `curl`. Siguen en `deny`: todo push a `main`, forzado o de borrado,
  `git merge`, `git reset --hard`, `gh release`, `wget` y `rm -rf`.
- El gancho `pre-push` no cambia: sigue rechazando `main`, etiquetas, borrados y push forzado salvo con
  `KITLEGAL_PUSH_HUMANO=1`, que el workflow nunca exporta. Una orden con ese prefijo es `ask`: la ve la persona.
- Un agente no modifica sus propios permisos: el cambio de `.claude/settings.json` lo aplica la persona.

## Consecuencias

- La persona lee el informe, dice qué hacer y confirma cada orden; ya no necesita saber los comandos.
- La garantía del ADR 0018 se mantiene por el mismo mecanismo que ya protegía las pausas: en headless, lo que pide
  confirmación no se ejecuta.
- `gh api` en `ask` hace que también las consultas de solo lectura pidan confirmación en una sesión interactiva; es
  el precio de no poder filtrar por método con un prefijo.
