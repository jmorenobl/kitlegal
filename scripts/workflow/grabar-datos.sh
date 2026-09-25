#!/usr/bin/env bash
# Grabación determinista de datos externos del workflow `hito` (ADR 0018, opción b).
#
#   scripts/workflow/grabar-datos.sh comprobar       # verificación de una tarea [datos] (verificar.sh)
#   scripts/workflow/grabar-datos.sh grabar <hito>   # tras el commit de una tarea [datos]
#
# El ejecutor nunca toca la red. Cuando una tarea [datos] añade o cambia un
# manifiesto <paquete>/testdata/grabaciones.json, el workflow graba después del
# commit de la tarea, sin modelo, y solo así:
#
#   1. La fuente del manifiesto (`fuente`) tiene fila en docs/SOURCES.md de
#      `main` con «Revisado» distinto de `pendiente`: los términos de uso, la
#      licencia y el robots.txt los revisó una persona ANTES del run.
#   2. Ejecuta el test de grabación del paquete (//go:build grabacion, función
#      TestGrabar*), que pide con internal/httpx —robots.txt, User-Agent y ritmo
#      de la fuente— con KITLEGAL_RECORD=1. Las respuestas que reproducen los
#      tests van a testdata/ del paquete; el material de origen del que se derivan
#      ficheros de data/ o verificaciones (hojas del INE, volcados del REL…) va a
#      $KITLEGAL_EVIDENCIAS = evidencias/<hito>/.
#   3. Solo se admiten ficheros nuevos o cambiados bajo testdata/ del paquete y
#      evidencias/<hito>/; cualquier otro se revierte. Cada fichero de evidencia
#      entra en evidencias/<hito>/manifiesto.json con su huella; los de más de
#      KITLEGAL_EVIDENCIA_MAX_BYTES (20 MiB) no se versionan: queda la huella y el
#      fichero se guarda en ~/.local/share/kitlegal/evidencias/<hito>/.
#   4. Registra la grabación en gates/grabaciones.md y la commitea. La persona la
#      revisa al final, en el informe del hito, antes de fusionar.
#
# `comprobar` exige 1 y el test de 2 ANTES del commit, como parte de la
# verificación de la tarea: un manifiesto sin fuente, con una fuente que nadie ha
# revisado o sin test de grabación es un defecto de la tarea (rojo → reparar →
# cuarentena), no una parada del run. `grabar` solo se detiene, con un dossier,
# si la fuente no responde tras un reintento: dependencia externa inaccesible
# (causa mayor).
set -euo pipefail
cd "$(dirname "$0")/../.."
. scripts/workflow/comun.sh

accion="${1:?uso: grabar-datos.sh comprobar | grabar <hito>}"
d=$(feature_dir)
t="$d/gates/tarea-actual.json"
id=$(jq -r .id "$t"); base=$(jq -r .base "$t"); datos=$(jq -r .datos "$t")

dossier() { # $1 causa, $2 detalle
  {
    echo "# Dossier: $1"
    echo
    echo "Tarea $id, paso grabar_datos ($(date -u +%Y-%m-%dT%H:%M:%SZ))."
    echo
    echo "$2"
    echo
    echo "Al resolverlo, reanudar con scripts/hito.sh --resume <run_id>: el bucle vuelve a esta tarea y la graba."
  } > "$d/gates/dossier.md"
  echo "grabar_datos detenido ($1). Dossier en $d/gates/dossier.md" >&2
  exit 3
}

fuente_revisada() { # fila `| \`fuente\` | … | Revisado |` en main, con Revisado ≠ pendiente
  { git show main:docs/SOURCES.md 2>/dev/null || git show origin/main:docs/SOURCES.md 2>/dev/null; } | awk -F'|' -v f="\`$1\`" '
    /^\|/ && NF > 3 {
      c = $2; gsub(/^[ \t]+|[ \t]+$/, "", c); sub(/[ \t]*\[\^.*\]$/, "", c)
      r = $(NF - 1); gsub(/^[ \t]+|[ \t]+$/, "", r)
      if (c == f && r != "pendiente" && r != "") ok = 1
    }
    END { exit !ok }'
}

nada() { jq -n --arg m "$1" '{grabado:false, motivo:$m}'; exit 0; }

# Defectos de un manifiesto que la tarea tiene que arreglar; uno por línea.
defectos_manifiesto() {
  local m="$1" fuente pkg
  fuente=$(jq -r '.fuente // ""' "$m" 2>/dev/null || echo "")
  if [ -z "$fuente" ]; then echo "$m no declara \`fuente\`"; return; fi
  fuente_revisada "$fuente" || echo "la fuente \`$fuente\` de $m no tiene fila en docs/SOURCES.md de main con «Revisado» fechado: el ejecutor no puede usarla (queda fuera de alcance; una persona la registraría antes del run)"
  pkg=${m%/testdata/grabaciones.json}
  grep -lE '^//go:build .*grabacion' "$pkg"/*_test.go 2>/dev/null | xargs grep -l 'func TestGrabar' >/dev/null 2>&1 \
    || echo "el paquete $pkg tiene manifiesto pero ningún test //go:build grabacion con una función TestGrabar*"
}

[ "$datos" = true ] || { [ "$accion" = comprobar ] && exit 0; nada "la tarea no lleva [datos]"; }
es_sha "$base" || { echo "grabar_datos: base inválida" >&2; exit 2; }

if [ "$accion" = comprobar ]; then
  manifiestos=$(cambiados_desde "$base" | grep -E '(^|/)testdata/grabaciones\.json$' || true)
  malos=""
  for m in $manifiestos; do [ -f "$m" ] && malos="$malos$(defectos_manifiesto "$m")
"; done
  if [ -n "$(printf '%s' "$malos" | tr -d '[:space:]')" ]; then
    printf '%s' "$malos" | grep . | sed 's/^/grabación: /'; exit 1
  fi
  [ -z "$manifiestos" ] || echo "grabación: manifiestos listos para grabar tras el commit: $(printf '%s ' $manifiestos)"
  exit 0
fi

[ "$accion" = grabar ] || { echo "acción desconocida: $accion" >&2; exit 2; }
hito="${2:?uso: grabar-datos.sh grabar <hito>}"
registro="$d/gates/grabaciones.md"
log="$d/gates/grabaciones.log"
max_bytes="${KITLEGAL_EVIDENCIA_MAX_BYTES:-20971520}"
slug=$(printf '%s' "$hito" | tr '[:upper:]' '[:lower:]' | tr . -)
evid="evidencias/$slug"
fuera_repo="$HOME/.local/share/kitlegal/evidencias/$slug"
manifiestos=$(git diff --name-only "$base" HEAD | grep -E '(^|/)testdata/grabaciones\.json$' || true)
[ -n "$manifiestos" ] || nada "la tarea no añade ni cambia ningún manifiesto de grabación"

[ -f "$registro" ] || printf '# Grabaciones del hito %s\n\nLas escribe el paso grabar_datos, sin modelo. Cada fichero lleva su huella SHA-256.\n' "$hito" > "$registro"
mkdir -p "$evid"
fuentes=""
for m in $manifiestos; do
  fuente=$(jq -r '.fuente // ""' "$m")
  pkg=${m%/testdata/grabaciones.json}
  # Ya lo exigió la verificación de la tarea; si aun así falla, algo cambió entre medias.
  mal=$(defectos_manifiesto "$m")
  [ -z "$mal" ] || dossier "manifiesto no grabable" "$mal"
  antes=$(mktemp); git status --porcelain --untracked-files=all | sort > "$antes"
  ok=0
  for intento in 1 2; do
    echo "== grabación de $fuente ($pkg), intento $intento" >> "$log"
    if KITLEGAL_RECORD=1 KITLEGAL_EVIDENCIAS="$PWD/$evid" go test -tags=grabacion -count=1 -run '^TestGrabar' "./$pkg/" >> "$log" 2>&1; then ok=1; break; fi
    [ "$intento" -eq 2 ] || sleep 60
  done
  [ "$ok" -eq 1 ] || dossier "dependencia externa inaccesible" "La grabación de \`$fuente\` falló dos veces. Final del log ($log):
\`\`\`
$(tail -40 "$log")
\`\`\`"
  nuevos=$(git status --porcelain --untracked-files=all | sort | comm -13 "$antes" - | cut -c4-)
  rm -f "$antes"
  admitidos=""
  for f in $nuevos; do
    case "$f" in
      "$d"/*) ;; # el registro y el log de este mismo paso
      "$pkg"/testdata/*|"$evid"/*) admitidos="$admitidos $f";;
      *) echo "grabar_datos: $f no es testdata/ de $pkg ni $evid/; se revierte" >&2
         if git cat-file -e "HEAD:$f" 2>/dev/null; then git checkout -q HEAD -- "$f"; else rm -f -- "$f"; fi;;
    esac
  done
  {
    printf '\n## %s · `%s` (%s)\n\n| Fichero | Bytes | SHA-256 | Versionado |\n|---|---|---|---|\n' "$(date -u +%Y-%m-%d)" "$fuente" "$pkg"
    for f in $admitidos; do
      [ -f "$f" ] || continue
      [ "$f" != "$evid/manifiesto.json" ] || continue
      b=$(wc -c < "$f" | tr -d ' '); h=$(huella "$f"); v=sí
      case "$f" in "$evid"/*)
        if [ "$b" -gt "$max_bytes" ]; then
          mkdir -p "$fuera_repo"; mv "$f" "$fuera_repo/"; v="no (más de $max_bytes bytes; copia en $fuera_repo)"
        fi
        [ -f "$evid/manifiesto.json" ] || echo '{"ficheros":[]}' > "$evid/manifiesto.json"
        jq --arg f "${f#"$evid"/}" --arg h "$h" --argjson b "$b" --arg fu "$fuente" --arg fe "$(date -u +%Y-%m-%d)" --arg v "$v" \
          '.ficheros = ([.ficheros[] | select(.fichero != $f)] + [{fichero:$f, sha256:$h, bytes:$b, fuente:$fu, fecha:$fe, versionado:$v}])' \
          "$evid/manifiesto.json" > "$evid/manifiesto.json.tmp" && mv "$evid/manifiesto.json.tmp" "$evid/manifiesto.json";;
      esac
      printf '| `%s` | %s | `%s` | %s |\n' "$f" "$b" "$h" "$v"
    done
  } >> "$registro"
  fuentes="$fuentes $fuente"
done

rmdir "$evid" 2>/dev/null || true
git add -A -- "$registro" "$log" $(printf '%s\n' $manifiestos | sed 's#/grabaciones\.json$##') $( [ -d "$evid" ] && echo "$evid" )
if git diff --cached --quiet; then nada "la grabación no cambió ningún fichero"; fi
git commit -q -m "chore($hito): grabaciones de$fuentes" -m "Paso grabar_datos del workflow, sin modelo, tras $id. Registro en $registro."
jq -n --arg f "${fuentes# }" --arg c "$(git rev-parse --short HEAD)" '{grabado:true, fuentes:$f, commit:$c}'
