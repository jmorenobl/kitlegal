#!/usr/bin/env bash
# Retira del árbol las marcas de escritura vacías que deja Claude Code.
#
#   scripts/workflow/marcas.sh   # scripts/claude-modelo.sh, al abrir la sesión de un paso y tras
#                                #   cada herramienta (gancho PostToolUse); verificar.sh, antes de medir
#
# Claude Code crea `<directorio>/.claude/.cc-writes/` para escribir un fichero de forma
# atómica dentro del sandbox, y lo deja vacío al terminar. git no lo ve (lo ignora el
# `~/.config/git/ignore` de quien lanza el run), pero los tests que leen o copian un
# directorio de grabaciones sí, y con razón no admiten en él nada que no sea un fichero
# regular. En H9 (run 1bd01ed1) una marca que dejó el paso `plan` en
# internal/source/boe/testdata/boe.legislacion-consolidada/ puso `make ci` en rojo, y
# T001, T002 y T003 fueron a cuarentena sin que su trabajo tuviera nada que ver.
#
# Solo se retira lo que está vacío: la marca y, si se queda vacío, el `.claude/` que la
# contiene. Nada con contenido, nada versionado, y nunca dentro de .git ni de
# node_modules. No lee la entrada estándar (el gancho la recibe, pero no la necesita) y
# sale siempre con 0: una marca que no se puede retirar no detiene ni la sesión ni el paso.
cd "$(dirname "$0")/../.." || exit 0
find . \( -name .git -o -name node_modules \) -prune \
  -o -type d -name .cc-writes -path '*/.claude/.cc-writes' -empty -print 2>/dev/null \
  | while IFS= read -r marca; do
      rmdir "$marca" 2>/dev/null && rmdir "${marca%/.cc-writes}" 2>/dev/null
    done
exit 0
