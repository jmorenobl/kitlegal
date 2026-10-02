#!/usr/bin/env bash
# Ejecutable de Claude para los pasos del workflow `hito`
# (SPECKIT_INTEGRATION_CLAUDE_EXECUTABLE, lo exporta scripts/hito.sh).
#
# spec-kit solo pasa `--model <valor>` a `claude -p`. Este wrapper acepta
# `<modelo>@<esfuerzo>` y lo traduce a `--model <modelo> --effort <esfuerzo>`,
# de modo que cada rol del workflow fija modelo y esfuerzo con un único input:
#
#   --model fable@xhigh   →  --model fable --effort xhigh
#   --model opus          →  --model opus               (esfuerzo por defecto)
#
# Es además la única puerta por la que el workflow y scripts/paso.sh abren una
# sesión con modelo, así que aquí se pone lo que esa sesión no puede hacer
# (ADR 0032; al final del fichero).
#
# El valor llega a argv desde los inputs del workflow: se valida con una
# expresión estricta y cualquier otro formato aborta con exit 2. El prompt
# (argumento de -p) se copia sin interpretarlo.
set -euo pipefail

patron='^(fable|opus|sonnet|haiku|claude-[a-z0-9-]+)(@(low|medium|high|xhigh|max))?$'

args=()
ajustes_de_quien_llama=""
while [ $# -gt 0 ]; do
  case "$1" in
    -p|--print)
      args+=("$1"); shift
      if [ $# -gt 0 ]; then args+=("$1"); shift; fi
      ;;
    --model)
      [ $# -ge 2 ] || { echo "claude-modelo: --model sin valor" >&2; exit 2; }
      if [[ ! "$2" =~ $patron ]]; then
        echo "claude-modelo: modelo inválido '$2' (formato: <modelo>[@low|medium|high|xhigh|max])" >&2
        exit 2
      fi
      args+=(--model "${BASH_REMATCH[1]}")
      if [ -n "${BASH_REMATCH[3]}" ]; then args+=(--effort "${BASH_REMATCH[3]}"); fi
      shift 2
      ;;
    --settings)
      # Claude Code solo atiende un --settings, el último: el de quien llama se guarda y
      # los ajustes del paso se le ponen encima más abajo.
      [ $# -ge 2 ] || { echo "claude-modelo: --settings sin valor" >&2; exit 2; }
      ajustes_de_quien_llama="$2"; shift 2
      ;;
    *)
      args+=("$1"); shift
      ;;
  esac
done

# Las sesiones headless se reconocen en sus transcripts por el entrypoint
# (scripts/hito.sh «limite_api» y scripts/coste-run.sh). `claude -p` escribe
# `sdk-cli` salvo que CLAUDE_CODE_ENTRYPOINT venga del entorno, y la extensión de
# VS Code exporta `claude-vscode` a todo proceso hijo: un hito lanzado desde ella
# dejaba sesiones que ninguno de los dos scripts reconocía. Se fija aquí para que
# el transcript diga lo mismo se lance el hito desde donde se lance.
export CLAUDE_CODE_ENTRYPOINT=sdk-cli

# Lo que una sesión de un paso no puede hacer lo impide algo que no es su prompt
# (ADR 0032; docs/WORKFLOW.md «Lo que una sesión de un paso no puede hacer»). En H7.4
# `reparar_cierre` leyó de un fichero el token de la suscripción de quien lanzó el run
# y dejó en segundo plano un sondeo con modelo. Todo se pone aquí, en la única puerta
# por la que el workflow (scripts/hito.sh) y scripts/paso.sh abren una sesión:
#
#   · El sandbox de Claude Code (Seatbelt en macOS, bubblewrap en Linux) en toda orden
#     de Bash y en todo lo que esa orden ejecute, guiones y programas incluidos, sin la
#     salida `dangerouslyDisableSandbox`, y la sesión no arranca si no está disponible:
#       – red: solo los tres servidores de las herramientas de Go (módulos, sumas y
#         vulnerabilidades) y el equipo local, que usan los tests. Sin api.anthropic.com
#         no hay sesión con modelo que valga, se lance como se lance, y sin github.com
#         no hay push ni gh. En macOS, Go comprueba los certificados con el servicio del
#         sistema (trustd), y sin `enableWeakerNetworkIsolation` govulncheck no llega
#         a vuln.go.dev (x509: OSStatus -26276);
#       – escritura: el repositorio, el directorio temporal ($TMPDIR, que el sandbox
#         pone a la sesión, y el del usuario en macOS) y las cachés de Go —la de
#         compilación, la de módulos y la de la base de sumas— y de golangci-lint. Ni
#         /tmp ni el resto del equipo;
#       – lectura: todo menos los directorios con secretos: ~/.config/kitlegal,
#         ~/.config/gh, ~/.ssh y los de KITLEGAL_SECRETOS (rutas absolutas separadas
#         por «:»). Una regla `deny` de Read sola no basta: no cubre un `cat`.
#   · Las mismas rutas, denegadas a la herramienta Read, que no pasa por el sandbox.
#   · La marca con la que el gancho PreToolUse le aplica scripts/workflow/politica-paso.sh,
#     que cubre Write y Edit —tampoco pasan por el sandbox— y da el motivo de cada
#     denegación en la propia sesión.
#   · Sin segundo plano: Claude Code no ofrece `run_in_background` ni pasa una orden
#     al fondo cuando vence su plazo.
#   · El `claude` de su PATH, y del de todo lo que ejecute, es el de
#     scripts/workflow/sin-modelo/, que no abre ninguna sesión y dice por qué.
#   · No lee .claude/settings.local.json: los permisos y los directorios adicionales
#     de la sesión interactiva de quien lanza el run no llegan a las del run.
raiz="$(cd "$(dirname "$0")/.." && pwd -P)"
claude="$(command -v "${KITLEGAL_CLAUDE_BIN:-claude}")" \
  || { echo "claude-modelo: no encuentro '${KITLEGAL_CLAUDE_BIN:-claude}' en el PATH" >&2; exit 2; }

base='{}'
if [ -f "$ajustes_de_quien_llama" ]; then
  base="$(cat "$ajustes_de_quien_llama")"
elif [ -n "$ajustes_de_quien_llama" ]; then
  base="$ajustes_de_quien_llama"
fi
# El temporal del usuario en macOS (/var/folders/…/T) es donde escriben `mktemp` y los
# programas que no miran TMPDIR, como los guiones que ejecutan los tests.
case "$(uname -s)" in
  Darwin) caches="$HOME/Library/Caches"; temporal="$(getconf DARWIN_USER_TEMP_DIR 2>/dev/null || true)";;
  *) caches="${XDG_CACHE_HOME:-$HOME/.cache}"; temporal="";;
esac
# Donde Go guarda lo que ya ha comprobado de la base de sumas (sum.golang.org): está junto
# a la caché de módulos, no dentro, y sin poder escribir ahí `go get` y `go mod tidy` no
# verifican ningún módulo que la caché no tenga ya. En H21 eso fijó la versión de una
# dependencia por lo que había en la caché del equipo y no por la que el plan quería.
gopath="$(go env GOPATH 2>/dev/null || true)"; gopath="${gopath%%:*}"
gosumdb=""; [ -n "$gopath" ] && gosumdb="$gopath/pkg/sumdb"
ajustes="$(jq -cn --argjson base "$base" --arg home "$HOME" --arg secretos "${KITLEGAL_SECRETOS:-}" \
  --arg gocache "$(go env GOCACHE 2>/dev/null || true)" --arg gomodcache "$(go env GOMODCACHE 2>/dev/null || true)" \
  --arg gotelemetria "$(go env GOTELEMETRYDIR 2>/dev/null || true)" --arg gosumdb "$gosumdb" \
  --arg golangci "${GOLANGCI_LINT_CACHE:-$caches/golangci-lint}" --arg temporal "${temporal%/}" '
  ([$home + "/.config/kitlegal", $home + "/.config/gh", $home + "/.ssh"]
    + ($secretos | split(":") | map(select(startswith("/"))))) as $vedados
  | {
      sandbox: {
        enabled: true,
        failIfUnavailable: true,
        allowUnsandboxedCommands: false,
        autoAllowBashIfSandboxed: false,
        enableWeakerNetworkIsolation: true,
        filesystem: {
          allowWrite: ([$temporal, $gocache, $gomodcache, $gosumdb, $gotelemetria, $golangci] | map(select(startswith("/")))),
          denyRead: $vedados
        },
        network: {
          allowedDomains: ["proxy.golang.org", "sum.golang.org", "vuln.go.dev"],
          allowLocalBinding: true
        }
      },
      permissions: {deny: ($vedados | map("Read(/" + . + "/**)"))}
    } as $paso
  | ($base * $paso)
  | .permissions.deny = (($base.permissions.deny // []) + $paso.permissions.deny)')" \
  || { echo "claude-modelo: no se han podido componer los ajustes de la sesión (¿--settings con un JSON que no vale?)" >&2; exit 2; }

export KITLEGAL_PASO_DE_WORKFLOW=1
export CLAUDE_CODE_DISABLE_BACKGROUND_TASKS=1
export PATH="$raiz/scripts/workflow/sin-modelo:$PATH"

exec "$claude" ${args[@]+"${args[@]}"} --setting-sources user,project --settings "$ajustes"
