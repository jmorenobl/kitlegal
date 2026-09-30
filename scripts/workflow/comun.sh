# Funciones comunes de los pasos shell del workflow `hito` (scripts/workflow/*.sh).
# Se carga con `. scripts/workflow/comun.sh` desde la raíz del repositorio, que es
# donde spec-kit ejecuta los pasos shell. Compatible con bash 3.2 (macOS).

# Directorio del feature del hito en curso (specs/NNN-hN-slug).
feature_dir() {
  jq -r .feature_directory .specify/feature.json
}

# Huella SHA-256 de un fichero, con la herramienta que haya en la máquina.
huella() {
  if command -v sha256sum >/dev/null 2>&1; then sha256sum "$1" | cut -d' ' -f1
  else shasum -a 256 "$1" | cut -d' ' -f1; fi
}

# Añade una línea con fecha a <feature_dir>/gates/supuestos.md, el registro de
# decisiones conservadoras y pendientes que el informe final enseña a la persona.
# Cada línea empieza por su impacto entre corchetes (ADR 0028); las que escribe un
# paso shell hablan del propio run —un gate agotado, una tarea en cuarentena, un
# cambio apartado— y llevan `[proceso]`.
anotar_supuesto() {
  local d s
  d=$(feature_dir); s="$d/gates/supuestos.md"
  mkdir -p "$d/gates"
  [ -f "$s" ] || printf '# Supuestos y pendientes del run\n\nCada línea la escribe un paso del workflow y empieza por su impacto: [comportamiento], [alcance], [skill], [interno] o [proceso]. El informe final las ordena por él.\n\n' > "$s"
  printf -- '- [proceso] %s · %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*" >> "$s"
}

# Directorio, dentro del git-dir del árbol de trabajo (uno por worktree), donde
# gate.sh guarda la instantánea de los artefactos que juzgó cada ronda. No se
# versiona: lo que se versiona es el diff de cada corrección (gates/*-correccion-r<n>.diff).
dir_instantaneas() {
  printf '%s/kitlegal-gates\n' "$(git rev-parse --git-dir)"
}

# Ficheros cambiados respecto a una base: modificados, añadidos y sin seguimiento.
cambiados_desde() {
  { git diff --name-only "$1"; git ls-files --others --exclude-standard; } | sort -u
}

# Lo que cambia el producto: todo lo de fuera de <feature_dir>/gates/, que son los
# registros del run (veredictos, supuestos, mediciones). Lo que el run cambia fuera
# de gates/ —código, tests, skill, evals, esquemas, datos, documentación o artefactos
# del feature— lo tiene que haber visto un juez de la revisión final antes de
# terminar, y una medición del cierre vale solo para el producto que midió (ADR 0030).
#
# Ficheros de fuera de gates/ que cambian entre el commit $1 y la cabeza, uno por línea.
fuera_de_gates_entre() {
  git diff --name-only "$1" HEAD -- . ":(exclude)$(feature_dir)/gates" | sort -u
}

# Ficheros de fuera de gates/ que tiene un commit, uno por línea.
fuera_de_gates_en() {
  git diff-tree --no-commit-id --name-only -r --root "$1" -- . ":(exclude)$(feature_dir)/gates"
}

# Cambios sin commitear fuera de gates/ (modificados, añadidos o sin seguimiento).
fuera_de_gates_sin_commitear() {
  git status --porcelain --untracked-files=all -- . ":(exclude)$(feature_dir)/gates" | cut -c4-
}

# ¿Es el árbol de ahora el mismo producto que el commit $1? Nada cambia fuera de
# gates/, ni en commits posteriores ni sin commitear.
producto_igual() {
  git cat-file -e "$1^{commit}" 2>/dev/null || return 1
  [ -z "$(fuera_de_gates_entre "$1")" ] && [ -z "$(fuera_de_gates_sin_commitear)" ]
}

# Rondas de la revisión final que emitieron veredicto, con la cabeza que juzgó
# cada una, una por línea: «<ronda> <ciclo> <sha> <juez A> <juez B>». Salen de
# gates/revision-juzgado.json, que escribe `gate.sh leer revision` desde el workflow
# 2.3.0. En un run anterior salen de los commits «docs(<hito>): veredictos de la
# revisión final»: cada uno cierra lo que juzgaron las rondas que versiona, y su
# propio contenido fuera de gates/ es el de la última cabeza juzgada.
rondas_juzgadas() { # $1 = hito
  local j v
  j="$(feature_dir)/gates/revision-juzgado.json"
  if [ -f "$j" ]; then
    jq -r '.rondas[] | "\(.ronda) \(.ciclo) \(.sha) \(.juez_a) \(.juez_b)"' "$j"
    return
  fi
  for v in $(git log --reverse --format=%H --fixed-strings --grep="docs($1): veredictos de la revisión final" main..HEAD 2>/dev/null); do
    printf '%s 1 %s %s %s\n' \
      "$(git show "$v:$(feature_dir)/gates/revision-rondas" 2>/dev/null || echo '?')" "$v" \
      "$(git show "$v:$(feature_dir)/gates/revision-a.json" 2>/dev/null | jq -r '.veredicto // "?"' 2>/dev/null || echo '?')" \
      "$(git show "$v:$(feature_dir)/gates/revision-b.json" 2>/dev/null | jq -r '.veredicto // "?"' 2>/dev/null || echo '?')"
  done
}

# La cabeza que juzgó el último veredicto de la revisión final, o nada si ninguno.
ultimo_juzgado() { # $1 = hito
  rondas_juzgadas "$1" | tail -1 | cut -d' ' -f3
}

# Aparta el trabajo hecho desde la base $1 —commits posteriores y cambios sin
# commitear, fuera del directorio del feature— como parche en $2 y devuelve el
# árbol y la rama a esa base. El directorio del feature (tasks.md, notas,
# veredictos) conserva su contenido actual. Es la vuelta a la última integración
# en verde: el trabajo apartado no se pierde, queda en el parche para la persona.
aparcar_desde() {
  local base="$1" parche="$2" d nuevos copia
  d=$(feature_dir)
  mkdir -p "$(dirname "$parche")"
  nuevos=$(git ls-files --others --exclude-standard | grep -v "^$d/" || true)
  [ -z "$nuevos" ] || git add -N -- $nuevos
  git diff --binary "$base" -- . ":(exclude)$d" > "$parche" || true
  [ -z "$nuevos" ] || git reset -q -- $nuevos
  copia=$(mktemp -d)
  cp -R "$d/." "$copia/"
  git reset -q --hard "$base"
  for f in $nuevos; do rm -f -- "$f"; done
  rm -rf "$d"; mkdir -p "$d"; cp -R "$copia/." "$d/"; rm -rf "$copia"
  # La suite de aceptación congelada vive en parte dentro del directorio del feature:
  # esos ficheros no se conservan como los demás, vuelven a su contenido de la base.
  if [ -f "$d/gates/aceptacion-congelada.json" ]; then
    for f in $(jq -r '.ficheros | keys[]' "$d/gates/aceptacion-congelada.json"); do
      git cat-file -e "$base:$f" 2>/dev/null && git checkout -q "$base" -- "$f"
    done
  fi
}

# ¿Es $1 un sha de commit completo?
es_sha() {
  printf '%s' "$1" | grep -qE '^[0-9a-f]{40}$'
}

# Ids de requisito que menciona una línea, con los rangos «FR-001 a FR-008» expandidos.
ids_de_linea() {
  printf '%s\n' "$1" | awk '{
    s = $0
    while (match(s, /(FR|SC)-[0-9]+ (a|al|–|-) (FR|SC)-[0-9]+/)) {
      r = substr(s, RSTART, RLENGTH); split(r, p, /[^A-Z0-9-]+/)
      pre = substr(p[1], 1, 3); a = substr(p[1], 4) + 0; b = substr(p[length(p)], 4) + 0
      if (b >= a && b - a < 200) for (i = a; i <= b; i++) printf "%s%03d\n", pre, i
      s = substr(s, RSTART + RLENGTH)
    }
    t = $0
    while (match(t, /(FR|SC)-[0-9]+/)) { print substr(t, RSTART, RLENGTH); t = substr(t, RSTART + RLENGTH) }
  }' | sort -u
}

# Filas de la sección «## Controles de umbral» de plan.md (ADR 0029): una por cada
# requisito con un umbral que se mide en make ci o en el cierre, con la forma
#   | Requisito | Umbral | Control | Dónde |
# Imprime, por fila, la celda de requisitos y la celda «Dónde» sin acentos graves,
# separadas por un tabulador. Solo se leen la primera y la última celda: un `|`
# dentro de las del medio no desplaza nada.
controles_de_umbral() { # $1 = plan.md
  [ -f "$1" ] || return 0
  awk '
    /^## / { dentro = ($0 ~ /^## Controles de umbral/); next }
    dentro && /^[[:space:]]*\|/ {
      n = split($0, c, "|"); req = c[2]; donde = c[n - 1]
      gsub(/`/, "", donde); gsub(/^[ \t]+|[ \t]+$/, "", req); gsub(/^[ \t]+|[ \t]+$/, "", donde)
      if (req ~ /^[ \t:-]*$/ || req == "Requisito") next # separador y cabecera
      print req "\t" donde
    }' "$1"
}

# Controles de una celda «Dónde»: `ci:<ruta>` o `ci:<ruta>:<Test>` (un test o una
# comprobación que ejecuta make ci; <Test> es una función Go o un objetivo del
# Makefile) y `evals:<skill>:<nombre>` (un umbral que el job de evals publica en
# `umbrales` de su informe.json). Uno por línea.
controles_de_celda() {
  printf '%s\n' "$1" | grep -oE '(ci:[A-Za-z0-9_./-]+(:[A-Za-z0-9_-]+)?|evals:[a-z0-9-]+:[A-Za-z0-9_.:-]+)' | sed -E 's/[.:]+$//' || true
}
