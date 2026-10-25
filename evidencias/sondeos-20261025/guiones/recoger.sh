#!/bin/bash
# Recoge el informe de boe-legislacion de cada ejecución de evals desde el 2026-10-05, la primera con juez (H24).
# Uso: recoger.sh <directorio de trabajo>, con <directorio>/runs.tsv escrito antes con
#   gh run list --workflow evals --limit 200 --json databaseId,headSha,headBranch,conclusion,createdAt,event \
#     --jq '.[]|[.databaseId,.headSha[0:8],.headBranch,.conclusion,.createdAt[0:16],.event]|@tsv'
# Deja el log de cada trabajo en <directorio>/logs y su informe.json en <directorio>/inf; tabla.py los resume.
set -euo pipefail
D="${1:?directorio de trabajo}"
G="$(cd "$(dirname "$0")" && pwd -P)"
cd "$(git rev-parse --show-toplevel)"
mkdir -p "$D/logs" "$D/inf"
awk -F'\t' '$4!="cancelled" && $4!="skipped" && $5>="2026-10-05"' "$D/runs.tsv" | while IFS=$'\t' read -r id sha _rama _concl fecha _ev; do
	[ -s "$D/inf/$id.json" ] && continue
	job=$(gh run view "$id" --json jobs --jq '.jobs[]|select(.name=="evals (boe-legislacion)" and .conclusion!="skipped")|.databaseId' 2> /dev/null | head -1)
	[ -z "$job" ] && { echo "$id sin trabajo"; continue; }
	gh run view "$id" --job "$job" --log 2> /dev/null | cut -f3- > "$D/logs/$id.log"
	if python3 -I "$G/extraer.py" "$D/logs/$id.log" "$D/inf/$id.json" 2> /dev/null; then
		echo "$id $sha $fecha ok"
	else
		echo "$id $sha $fecha sin informe"
		rm -f "$D/inf/$id.json"
	fi
done
