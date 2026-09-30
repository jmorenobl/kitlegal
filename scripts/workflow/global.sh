#!/usr/bin/env bash
# Pasos que tocan el hito entero, fuera del bucle de tareas: reparaciones de make ci,
# barrido de artefactos, correctores de la revisión final y reparaciones del cierre.
#
#   scripts/workflow/global.sh base [<mensaje>]                       # antes del paso con modelo: fija la base
#   scripts/workflow/global.sh cerrar <hito> [<mensaje>] [--revertir-si-rojo]   # después → JSON
#   scripts/workflow/global.sh retomar <hito>                         # cierra un paso interrumpido → JSON
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
#
# Con <mensaje>, base lo guarda en gates/base-global-pendiente hasta que cerrar
# termina, y cerrar lo usa si no recibe otro. Así, si el run se interrumpe a mitad
# del paso con modelo (un límite de uso, un corte), retomar —al volver a entrar en el
# bucle de la revisión y el cierre, scripts/workflow/revision.sh pendiente— hace el
# cerrar que faltaba: el trabajo a medias se verifica y se commitea o se aparta, y
# nunca queda sin commitear ni fuera de la vista de los jueces (ADR 0030).
set -euo pipefail
cd "$(dirname "$0")/../.."
. scripts/workflow/comun.sh

d=$(feature_dir); g="$d/gates"; mkdir -p "$g"
f_base="$g/base-global"
f_pendiente="$g/base-global-pendiente"

cerrar() { # $1 = hito, $2 = mensaje, $3 = --revertir-si-rojo o nada
  local hito="$1" mensaje="$2" revertir="${3:-}" base rc guardian n parche causa
  base=$(cat "$f_base"); es_sha "$base" || { echo "base global inválida" >&2; exit 2; }
  rc=0; scripts/workflow/verificar.sh global "$base" >/dev/null || rc=$?
  guardian=false; grep -q '^guardián: ' "$g/ci.log" && ! grep -q '^guardián de diff ok' "$g/ci.log" && guardian=true
  if [ "$rc" -ne 0 ] && { [ "$guardian" = true ] || [ "$revertir" = --revertir-si-rojo ]; }; then
    n=$(find "$g" -maxdepth 1 -name 'aparcado-*.patch' | wc -l | tr -d ' '); parche="$g/aparcado-$((n + 1)).patch"
    causa=$([ "$guardian" = true ] && echo "el guardián global lo rechazó" || echo "make ci quedó en rojo")
    grep '^guardián' "$g/ci.log" | head -5 >&2 || true
    aparcar_desde "$base" "$parche"
    anotar_supuesto "«${mensaje}»: $causa; el cambio se apartó en $parche y el hito vuelve a $(printf '%s' "$base" | cut -c1-7)."
    rm -f "$f_pendiente"
    jq -n --arg p "$parche" --arg c "$causa" '{verde:false, apartado:true, parche:$p, causa:$c}'
    return 0
  fi
  # Sin cambios fuera de gates/ no hay nada que commitear con este mensaje: los
  # registros del run (gates/base-global, supuestos, veredictos) los versiona el paso
  # siguiente que commitee. Los artefactos del feature (spec.md, plan.md,
  # quickstart.md, research.md, tasks.md…) sí son una corrección: en H7.2 la de la
  # primera ronda de la revisión solo tocó quickstart.md y research.md, quedó sin
  # commitear y la arrastró `publicar` como «registros del run», después de los
  # veredictos; el informe la listó como un commit que ningún juez vio (ADR 0029).
  if [ -z "$(fuera_de_gates_sin_commitear)" ]; then
    rm -f "$f_pendiente"
    jq -n --argjson v "$([ "$rc" -eq 0 ] && echo true || echo false)" '{verde:$v, apartado:false, sin_cambios:true}'
    return 0
  fi
  git add -A
  if ! git diff --cached --quiet; then
    if [ "$rc" -eq 0 ]; then git commit -q -m "$mensaje"
    else git commit -q -m "$mensaje (make ci en rojo)" -m "Última verificación: $g/ci.log."; fi
  fi
  [ "$rc" -eq 0 ] || anotar_supuesto "«${mensaje}»: make ci sigue en rojo; lo verán el paso siguiente y el informe final."
  rm -f "$f_pendiente"
  jq -n --argjson v "$([ "$rc" -eq 0 ] && echo true || echo false)" '{verde:$v, apartado:false}'
}

case "${1:?uso: global.sh base [<mensaje>] | cerrar <hito> [<mensaje>] [--revertir-si-rojo] | retomar <hito>}" in
  base)
    git rev-parse HEAD > "$f_base"
    if [ -n "${2:-}" ]; then printf '%s\n' "$2" > "$f_pendiente"; else rm -f "$f_pendiente"; fi
    echo "base global: $(cut -c1-7 "$f_base")";;

  cerrar)
    hito="${2:?uso: global.sh cerrar <hito> [<mensaje>] [--revertir-si-rojo]}"
    mensaje="${3:-}"; revertir="${4:-}"
    if [ "$mensaje" = --revertir-si-rojo ]; then mensaje=""; revertir=--revertir-si-rojo; fi
    [ -n "$mensaje" ] || mensaje=$(cat "$f_pendiente" 2>/dev/null || true)
    [ -n "$mensaje" ] || { echo "global.sh cerrar: sin mensaje (ni argumento ni el que guardó base)" >&2; exit 2; }
    cerrar "$hito" "$mensaje" "$revertir";;

  retomar)
    hito="${2:?uso: global.sh retomar <hito>}"
    if [ ! -f "$f_pendiente" ]; then jq -n '{retomado:false}'; exit 0; fi
    mensaje=$(cat "$f_pendiente")
    echo "retomar: el paso de «${mensaje}» se interrumpió antes de cerrarse; se cierra ahora" >&2
    r=$(cerrar "$hito" "$mensaje" "")
    if [ "$(jq -r '.sin_cambios // false' <<<"$r")" = true ]; then
      anotar_supuesto "«${mensaje}»: el run se interrumpió antes de cerrar el paso, que no había dejado cambios fuera de gates/; el bucle lo vuelve a ejecutar."
    else
      anotar_supuesto "«${mensaje}»: el run se interrumpió antes de cerrar el paso; al reanudar se verificó lo que había dejado ($(jq -r 'if .apartado then "apartado en " + .parche elif .verde then "commiteado, make ci en verde" else "commiteado con make ci en rojo" end' <<<"$r")), y lo juzgará la revisión final."
    fi
    jq -c '. + {retomado:true}' <<<"$r";;

  *) echo "subcomando desconocido: $1" >&2; exit 2;;
esac
