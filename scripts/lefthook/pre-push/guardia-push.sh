#!/usr/bin/env bash
# Guardia de push (gancho `pre-push` de lefthook, job `script` con `use_stdin`;
# lefthook.yml fija source_dir: scripts/lefthook).
#
# Capa mecánica de la política del workflow `hito`: el workflow y las tareas
# [plataforma] empujan la rama del hito y abren la propuesta de cambio; fusionar
# en `main` y publicar releases son acciones humanas (constitución, «Flujo de
# trabajo y gates» §4). Los permisos de .claude/settings.json filtran las órdenes
# por prefijo; este gancho ve lo que git va a enviar de verdad, sea cual sea la
# forma de la orden.
#
# Rechaza: cualquier actualización de `main`, borrados de referencias remotas,
# actualizaciones que no avanzan en línea recta (push forzado o rama divergente)
# y etiquetas. Una persona puede saltarlo con KITLEGAL_PUSH_HUMANO=1 (p. ej. para
# etiquetar una release); el workflow nunca exporta esa variable.
#
# git entrega por stdin una línea por referencia:
#   <ref local> <sha local> <ref remota> <sha remoto>
set -euo pipefail

[ "${KITLEGAL_PUSH_HUMANO:-0}" != 1 ] || exit 0

cero=0000000000000000000000000000000000000000
rc=0
while read -r ref_local sha_local ref_remota sha_remoto; do
  [ -n "$ref_remota" ] || continue
  case "$ref_remota" in
    refs/heads/main|refs/heads/master)
      echo "pre-push: actualizar '$ref_remota' es una acción humana (KITLEGAL_PUSH_HUMANO=1 para hacerlo a mano)" >&2; rc=1;;
    refs/tags/*)
      echo "pre-push: las etiquetas son del release, acción humana (KITLEGAL_PUSH_HUMANO=1)" >&2; rc=1;;
  esac
  if [ "$sha_local" = "$cero" ]; then
    echo "pre-push: borrar '$ref_remota' está prohibido" >&2; rc=1; continue
  fi
  if [ "$sha_remoto" != "$cero" ] && ! git merge-base --is-ancestor "$sha_remoto" "$sha_local" 2>/dev/null; then
    echo "pre-push: '$ref_remota' no avanza en línea recta desde lo publicado (push forzado o rama divergente); prohibido" >&2; rc=1
  fi
done
exit "$rc"
