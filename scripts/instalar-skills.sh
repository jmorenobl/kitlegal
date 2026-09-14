#!/usr/bin/env bash
# Enlaza cada skill de skills/ en el directorio personal de skills de Claude Code y deja bin/instalado/kitlegal
# apuntando al binario instalado que recibe, que es lo que resuelve el scripts/ de cada skill enlazada. Ante una
# entrada ajena con el nombre de una skill no crea ni cambia nada. Lo ejecuta make install tras go install, con la
# ruta que da go list -f '{{.Target}}' (contracts/instalacion.md §2 de H5).
set -euo pipefail

# 1. La raíz física del repositorio, que es el destino de los enlaces de las skills.
raiz="$(cd "$(dirname "$0")/.." && pwd -P)"

# 2. El binario instalado.
if [[ $# -ne 1 ]]; then
	echo "instalar-skills: uso: scripts/instalar-skills.sh <binario instalado>" >&2
	exit 1
fi

binario="$1"

if [[ ! -f "$binario" || ! -x "$binario" ]]; then
	echo "instalar-skills: $binario no existe o no es ejecutable" >&2
	exit 1
fi

# 3. El directorio personal.
if [[ -z "${HOME:-}" ]]; then
	echo "instalar-skills: HOME no está definido" >&2
	exit 1
fi

personal="$HOME/.claude/skills"

# 4. Cada entrada del directorio personal con el nombre de una skill: si no existe, se creará; si es un enlace cuyo
# destino literal es la skill, se deja; cualquier otra cosa —un directorio, un fichero, un enlace a otro sitio o un
# enlace roto— es un conflicto (data-model §11.1). readlink sin -f compara el destino literal y es igual en macOS y
# en Linux.
shopt -s nullglob
conflictos=0

for carpeta in "$raiz"/skills/*/; do
	nombre="$(basename "$carpeta")"
	destino="$raiz/skills/$nombre"
	entrada="$personal/$nombre"

	if [[ -L "$entrada" ]]; then
		if [[ "$(readlink "$entrada")" == "$destino" ]]; then
			continue
		fi
	elif [[ ! -e "$entrada" ]]; then
		continue
	fi

	echo "instalar-skills: conflicto: $entrada ya existe y no es un enlace a $destino; no se modifica" >&2
	conflictos=$((conflictos + 1))
done

# 5. Con un solo conflicto, nada se crea ni se cambia.
if [[ $conflictos -gt 0 ]]; then
	exit 1
fi

# 6. El enlace del binario instalado, que el scripts/ de cada skill alcanza por ../../../bin/instalado/kitlegal: se
# rehace igual en cada instalación (data-model §11.2).
mkdir -p "$raiz/bin/instalado"
ln -sfn "$binario" "$raiz/bin/instalado/kitlegal"

# 7 y 8. Los enlaces que faltan; tras el paso 4, toda entrada que ya es un enlace nombra su skill y se deja como está.
# Una línea por skill y otra para el binario.
mkdir -p "$personal"

for carpeta in "$raiz"/skills/*/; do
	nombre="$(basename "$carpeta")"

	if [[ ! -L "$personal/$nombre" ]]; then
		ln -s "$raiz/skills/$nombre" "$personal/$nombre"
	fi

	echo "instalar-skills: $nombre → $raiz/skills/$nombre"
done

echo "instalar-skills: kitlegal → $binario"
