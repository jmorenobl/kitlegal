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
anotar_supuesto() {
  local d s
  d=$(feature_dir); s="$d/gates/supuestos.md"
  mkdir -p "$d/gates"
  [ -f "$s" ] || printf '# Supuestos y pendientes del run\n\nCada línea la escribe un paso del workflow; el informe final las reúne.\n\n' > "$s"
  printf -- '- %s · %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*" >> "$s"
}

# Ficheros cambiados respecto a una base: modificados, añadidos y sin seguimiento.
cambiados_desde() {
  { git diff --name-only "$1"; git ls-files --others --exclude-standard; } | sort -u
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
