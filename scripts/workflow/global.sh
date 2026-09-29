#!/usr/bin/env bash
# Pasos que tocan el hito entero, fuera del bucle de tareas: reparaciones de make ci,
# barrido de artefactos, correctores de la revisión final y reparaciones del cierre.
#
#   scripts/workflow/global.sh base                       # antes del paso con modelo: fija la base
#   scripts/workflow/global.sh cerrar <hito> <mensaje> [--revertir-si-rojo]   # después → JSON
#
# cerrar verifica desde la base (guardián global + make ci, scripts/workflow/verificar.sh):
#   · en verde, commitea lo pendiente con <mensaje>;
#   · si el guardián lo rechaza (fixtures, esquemas, evidencias, configuración de
#     verificación o la suite de aceptación congelada), aparta todo lo hecho desde la
#     base como parche y vuelve a ella: es la vuelta a la última integración en verde;
#   · con make ci en rojo, commitea marcándolo en el mensaje —el paso siguiente (otra
#     reparación, el juez) parte de ahí y lo ve—, salvo con --revertir-si-rojo, que
#     aparta el cambio como en el caso anterior.
# Nunca para el run: lo que queda sin arreglar va al registro de supuestos.
set -euo pipefail
cd "$(dirname "$0")/../.."
. scripts/workflow/comun.sh

d=$(feature_dir); g="$d/gates"; mkdir -p "$g"
f_base="$g/base-global"

case "${1:?uso: global.sh base | cerrar <hito> <mensaje> [--revertir-si-rojo]}" in
  base)
    git rev-parse HEAD > "$f_base"; echo "base global: $(cut -c1-7 "$f_base")";;

  cerrar)
    hito="${2:?uso: global.sh cerrar <hito> <mensaje>}"; mensaje="${3:?uso: global.sh cerrar <hito> <mensaje>}"
    revertir="${4:-}"
    base=$(cat "$f_base"); es_sha "$base" || { echo "base global inválida" >&2; exit 2; }
    rc=0; scripts/workflow/verificar.sh global "$base" >/dev/null || rc=$?
    guardian=false; grep -q '^guardián: ' "$g/ci.log" && ! grep -q '^guardián de diff ok' "$g/ci.log" && guardian=true
    if [ "$rc" -ne 0 ] && { [ "$guardian" = true ] || [ "$revertir" = --revertir-si-rojo ]; }; then
      n=$(find "$g" -maxdepth 1 -name 'aparcado-*.patch' | wc -l | tr -d ' '); parche="$g/aparcado-$((n + 1)).patch"
      causa=$([ "$guardian" = true ] && echo "el guardián global lo rechazó" || echo "make ci quedó en rojo")
      grep '^guardián' "$g/ci.log" | head -5 >&2 || true
      aparcar_desde "$base" "$parche"
      anotar_supuesto "«${mensaje}»: $causa; el cambio se apartó en $parche y el hito vuelve a $(printf '%s' "$base" | cut -c1-7)."
      jq -n --arg p "$parche" --arg c "$causa" '{verde:false, apartado:true, parche:$p, causa:$c}'
      exit 0
    fi
    # Sin cambios fuera de gates/ no hay nada que commitear con este mensaje: los
    # registros del run (gates/base-global, supuestos, veredictos) los versiona el paso
    # siguiente que commitee. Los artefactos del feature (spec.md, plan.md,
    # quickstart.md, research.md, tasks.md…) sí son una corrección: en H7.2 la de la
    # primera ronda de la revisión solo tocó quickstart.md y research.md, quedó sin
    # commitear y la arrastró `publicar` como «registros del run», después de los
    # veredictos; el informe la listó como un commit que ningún juez vio (ADR 0029).
    if [ -z "$(git status --porcelain -- . ":(exclude)$d/gates")" ]; then
      jq -n --argjson v "$([ "$rc" -eq 0 ] && echo true || echo false)" '{verde:$v, apartado:false, sin_cambios:true}'
      exit 0
    fi
    git add -A
    if ! git diff --cached --quiet; then
      if [ "$rc" -eq 0 ]; then git commit -q -m "$mensaje"
      else git commit -q -m "$mensaje (make ci en rojo)" -m "Última verificación: $g/ci.log."; fi
    fi
    [ "$rc" -eq 0 ] || anotar_supuesto "«${mensaje}»: make ci sigue en rojo; lo verán el paso siguiente y el informe final."
    jq -n --argjson v "$([ "$rc" -eq 0 ] && echo true || echo false)" '{verde:$v, apartado:false}';;

  *) echo "subcomando desconocido: $1" >&2; exit 2;;
esac
