#!/usr/bin/env bash
# Guardián de diff del workflow `hito` (docs/WORKFLOW.md «Guardián de diff»).
#
#   scripts/workflow/guardian-diff.sh tarea           # la tarea de gates/tarea-actual.json
#   scripts/workflow/guardian-diff.sh global <base>   # reparaciones, barrido, correctores y cierre
#
# Modo tarea: el diff desde la base de la tarea se limita a las rutas que declara
# su línea de tasks.md (más go.mod, go.sum, CHANGELOG.md, el directorio del
# feature y x_test.go de un x.go declarado); testdata/ y schemas/ solo con
# [datos]; .golangci.yml sin declararlo solo si añade entradas a
# misspell.ignore-rules.
#
# Modo global: sin rutas declaradas, pero nadie toca testdata/, schemas/ ni la
# configuración de verificación (lint, cobertura, CI, ganchos, el propio
# workflow).
#
# En los dos modos: evidencias/ solo la escribe el paso grabar_datos; la suite
# de aceptación congelada (gates/aceptacion-congelada.json) no cambia; y ningún
# diff añade un t.Skip. Sale con 1 y explica cada violación; con 0 si todo está
# dentro. Lo ejecuta scripts/workflow/verificar.sh antes de make ci, de modo que
# una violación es un rojo más que el reparador ve, no una parada del run.
set -euo pipefail
cd "$(dirname "$0")/../.."
. scripts/workflow/comun.sh

modo="${1:?uso: guardian-diff.sh tarea | global <base>}"
d=$(feature_dir)

case "$modo" in
  tarea)
    t="$d/gates/tarea-actual.json"
    id=$(jq -r .id "$t"); base=$(jq -r .base "$t"); datos=$(jq -r .datos "$t")
    rutas=$(jq -r '.rutas[]' "$t");;
  global)
    base="${2:?uso: guardian-diff.sh global <base>}"; id="(global)"; datos=false; rutas="";;
  *) echo "modo desconocido: $modo" >&2; exit 2;;
esac
es_sha "$base" || { echo "guardián: base inválida '$base'" >&2; exit 2; }

cambiados=$(cambiados_desde "$base")
[ -n "$cambiados" ] || { echo "guardián: $id no cambió ningún fichero"; exit 0; }

# .golangci.yml se admite sin declararlo solo si el diff se limita a AÑADIR entradas a
# misspell.ignore-rules (palabras españolas que el diccionario inglés toma por erratas;
# en H3, T008 paró el run por una que no podía declarar). Se compara el fichero sin
# comentarios, líneas vacías ni ítems de ignore-rules: cualquier otra diferencia, o una
# entrada retirada, sigue exigiendo que la tarea declare .golangci.yml.
esqueleto() { awk '
  /^[[:space:]]*#/ || /^[[:space:]]*$/ { next }
  { match($0, /^[[:space:]]*/); ind = RLENGTH }
  en && ind <= ind_ir { en = 0 }
  en && $0 ~ /^[[:space:]]*- / { next }
  $0 ~ /^[[:space:]]*ignore-rules:[[:space:]]*$/ { en = 1; ind_ir = ind }
  { print }'; }
entradas() { awk '
  /^[[:space:]]*#/ || /^[[:space:]]*$/ { next }
  { match($0, /^[[:space:]]*/); ind = RLENGTH }
  en && ind <= ind_ir { en = 0 }
  en && $0 ~ /^[[:space:]]*- / { sub(/^[[:space:]]*- /, ""); print; next }
  $0 ~ /^[[:space:]]*ignore-rules:[[:space:]]*$/ { en = 1; ind_ir = ind }'; }
solo_ignore_rules() {
  local viejo nuevo nuevas e
  viejo=$(git show "$base:.golangci.yml" 2>/dev/null) || return 1
  nuevo=$(cat .golangci.yml 2>/dev/null) || return 1
  [ "$(printf '%s\n' "$viejo" | esqueleto)" = "$(printf '%s\n' "$nuevo" | esqueleto)" ] || return 1
  nuevas=$(printf '%s\n' "$nuevo" | entradas)
  for e in $(printf '%s\n' "$viejo" | entradas); do printf '%s\n' "$nuevas" | grep -qxF -- "$e" || return 1; done
}

fuera=""; protegidos=""; evidencia=""; config=""
for f in $cambiados; do
  case "$f" in go.mod|go.sum|CHANGELOG.md|"$d"/*) continue;; esac
  if [ "$f" = .golangci.yml ] && solo_ignore_rules; then echo "guardián: .golangci.yml solo añade entradas a misspell.ignore-rules; permitido"; continue; fi
  case "$f" in evidencias/*) evidencia="$evidencia $f"; continue;; esac
  case "$f" in testdata/*|*/testdata/*|schemas/*) protegidos="$protegidos $f";; esac
  if [ "$modo" = global ]; then
    case "$f" in
      .golangci.yml|codecov.yml|lefthook.yml|.github/*|tools/*|scripts/workflow/*|scripts/hito.sh|scripts/lefthook/*|.specify/*|.claude/*)
        config="$config $f";;
    esac
    continue
  fi
  ok=0
  for r in $rutas; do
    r=${r%%\**}; r=${r%/}
    [ -n "$r" ] || continue
    case "$f" in "$r"|"$r"/*) ok=1; break;; esac
    case "$r" in *.go) case "$f" in "${r%.go}_test.go") ok=1; break;; esac;; esac
  done
  [ "$ok" -eq 1 ] || fuera="$fuera $f"
done

rc=0
if [ -n "$protegidos" ] && [ "$datos" != "true" ]; then
  echo "guardián: $id modificó fixtures o esquemas sin etiqueta [datos]:$protegidos"; rc=1
fi
if [ -n "$evidencia" ]; then
  echo "guardián: $id escribió en evidencias/, que solo escribe el paso grabar_datos:$evidencia"; rc=1
fi
if [ -n "$config" ]; then
  echo "guardián: $id tocó la configuración de verificación, que ningún reparador ni corrector cambia:$config"; rc=1
fi
if [ -n "$fuera" ]; then
  echo "guardián: $id tocó ficheros fuera de sus rutas declaradas:$fuera"
  echo "rutas declaradas: $(printf '%s ' $rutas)"; rc=1
fi

# La suite de aceptación congelada no cambia después de congelarse (ADR 0018).
congelada="$d/gates/aceptacion-congelada.json"
if [ -f "$congelada" ]; then
  alterados=""
  for par in $(jq -r '.ficheros | to_entries[] | "\(.key)=\(.value)"' "$congelada"); do
    f=${par%%=*}; h=${par#*=}
    if [ ! -f "$f" ] || [ "$(huella "$f")" != "$h" ]; then alterados="$alterados $f"; fi
  done
  if [ -n "$alterados" ]; then
    echo "guardián: $id alteró la suite de aceptación congelada; se arregla el código, nunca el test:$alterados"; rc=1
  fi
fi

# Ningún diff añade un t.Skip: un test saltado no prueba nada (constitución, criterio 1).
saltos=$(git diff "$base" -U0 -- '*.go' | grep -E '^\+[^+].*\bt\.Skip(Now|f)?\(' || true)
nuevos_go=$(git ls-files --others --exclude-standard -- '*.go')
for f in $nuevos_go; do
  s=$(grep -nE '\bt\.Skip(Now|f)?\(' "$f" || true)
  [ -z "$s" ] || saltos="$saltos
$f: $s"
done
if [ -n "$(printf '%s' "$saltos" | tr -d '[:space:]')" ]; then
  echo "guardián: $id añade tests saltados:"; printf '%s\n' "$saltos"; rc=1
fi

[ "$rc" -eq 0 ] && echo "guardián de diff ok ($id, $(printf '%s\n' "$cambiados" | wc -l | tr -d ' ') ficheros)"
exit "$rc"
