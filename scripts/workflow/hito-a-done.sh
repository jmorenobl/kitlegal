#!/usr/bin/env bash
# Pasa el fichero del hito de docs/roadmap/in-progress/ (o de todo/) a done/ en el primer commit de su rama
# (docs/roadmap/README.md). Al fusionar la propuesta de cambio, main lo tiene hecho, y como el movimiento es parte del
# producto desde el principio lo ven todas las revisiones y todas las mediciones (ADR 0030). En main sigue en
# in-progress/ mientras el run no se fusiona.
#
#   scripts/workflow/hito-a-done.sh <hito>
#
# Al reanudar, un hito que ya está en done/ no se vuelve a mover. En main no mueve nada: el run trabaja en la rama del
# hito.
set -euo pipefail
cd "$(dirname "$0")/../.."
. scripts/workflow/comun.sh

hito="${1:?uso: hito-a-done.sh <hito>}"
f=$(fichero_del_hito "$hito") || { echo "el hito $hito no tiene fichero en docs/roadmap/" >&2; exit 1; }

mensaje="docs($hito): el hito pasa a done/ con su propuesta de cambio"

# En done/ y en la cabeza: un hito que ya pasó no se vuelve a mover.
case "$f" in
  docs/roadmap/done/*)
    if git cat-file -e "HEAD:$f" 2>/dev/null; then echo "el hito $hito ya está en $f"; exit 0; fi;;
esac
if [ "$(git branch --show-current)" = main ]; then
  echo "en main el hito no pasa a done/: se mueve en la rama del hito" >&2
  exit 1
fi

case "$f" in
  docs/roadmap/done/*)
    # En done/ pero no en la cabeza: el `git mv` de un intento anterior cuyo commit no terminó (en H41, el gancho de
    # pre-commit pasó del timeout del paso con la caché de lint fría tras unas actualizaciones de dependencias). Se
    # confirma el traslado, con el borrado del origen que dejó preparado, en vez de darlo por hecho.
    hito_re=$(printf '%s' "$hito" | sed 's/\./\\./g')
    origen=$(git diff --cached --no-renames --name-only --diff-filter=D -- docs/roadmap/in-progress docs/roadmap/todo |
      grep -E "/([0-9]{3}-)?${hito_re}-[^/]*\.md\$" || true)
    git add -- "$f"
    # shellcheck disable=SC2086 # $origen es una ruta del roadmap, sin espacios, o nada
    git commit -q -m "$mensaje" -- $origen "$f"
    echo "el hito $hito pasa a $f (traslado de un intento anterior sin confirmar)"
    exit 0;;
esac

destino="docs/roadmap/done/$(basename "$f" | sed -E 's/^[0-9]{3}-//')"
mkdir -p "$(dirname "$destino")"
git mv "$f" "$destino"
git commit -q -m "$mensaje" -- "$f" "$destino"
echo "el hito $hito pasa a $destino"
