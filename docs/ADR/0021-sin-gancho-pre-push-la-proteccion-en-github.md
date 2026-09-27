# 0021 · Sin gancho `pre-push`: la protección vive en GitHub y en los permisos

- **Estado**: aceptada
- **Fecha**: 2026-09-27
- **Hito**: transversal (constitución 2.3.0, tras publicar v0.1.1). Sustituye la parte del ADR 0007 que confiaba al
  gancho `pre-push` la política de publicación, y la del ADR 0020 que lo mantenía.

## Contexto y problema

El ADR 0007 puso un gancho `pre-push` (`scripts/lefthook/pre-push/guardia-push.sh`) como barrera mecánica: rechazaba
`main`, push forzado, borrados y etiquetas, salvo con `KITLEGAL_PUSH_HUMANO=1`. Era la única barrera en el servidor
que podía tenerse entonces: el repositorio era privado en el plan gratuito y `main` no admitía protección.

Desde el 2026-09-27 el repositorio es público y `main` está protegida en GitHub: exige `ci` y `snapshot` en verde,
ramas al día, y prohíbe el push forzado y el borrado. Y desde el ADR 0020, fusionar y empujar una etiqueta son `ask`
en `.claude/settings.json`: en una sesión interactiva los confirma la persona y en headless no se ejecutan. El gancho
ya no protege nada que no esté protegido, y obliga a anteponer `KITLEGAL_PUSH_HUMANO=1` a cada etiqueta, un prefijo
que además el clasificador de Claude Code lee como un intento de saltarse una salvaguarda.

## Decisión

Se retira el gancho `pre-push` (la sección de `lefthook.yml` y el guion). Lo que el run no puede hacer queda en dos
sitios:

- **Servidor**: la protección de `main` en GitHub (checks `ci` y `snapshot`, `strict`, sin push forzado ni borrado).
  Vale para cualquier cliente, no solo para el clon con los ganchos instalados.
- **Agente**: `.claude/settings.json`. `deny` para todo push a `main`, forzado o de borrado, `git merge`,
  `gh release`; `ask` para `gh pr merge`, `gh repo edit`, `gh api`, `git push origin v…` (antes con el prefijo
  `KITLEGAL_PUSH_HUMANO=1`, que desaparece) y `curl`.

`KITLEGAL_PUSH_HUMANO` deja de existir. Los ganchos `pre-commit` no cambian.

## Consecuencias

- Etiquetar una release es `git push origin v<x.y.z>`, confirmado por la persona.
- Las etiquetas no tienen protección en el servidor: la única barrera frente al run es el `ask` de los permisos (en
  headless no se ejecuta). Si hiciera falta más, un conjunto de reglas de etiquetas en GitHub sería el sitio.
- Quien tenga el gancho instalado de antes lo retira con `rm .git/hooks/pre-push` (o `make hooks`, que reinstala
  solo los de la configuración).
