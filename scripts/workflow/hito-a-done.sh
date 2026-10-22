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

case "$f" in
  docs/roadmap/done/*) echo "el hito $hito ya está en $f"; exit 0;;
esac
if [ "$(git branch --show-current)" = main ]; then
  echo "en main el hito no pasa a done/: se mueve en la rama del hito" >&2
  exit 1
fi

destino="docs/roadmap/done/$(basename "$f" | sed -E 's/^[0-9]{3}-//')"
mkdir -p "$(dirname "$destino")"
git mv "$f" "$destino"
git commit -q -m "docs($hito): el hito pasa a done/ con su propuesta de cambio" -- "$f" "$destino"
echo "el hito $hito pasa a $destino"
