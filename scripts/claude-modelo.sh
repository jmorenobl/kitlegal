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
# Detrás puede ir `#<paso>:<rol>`, que cada paso prompt de workflow.yml pone a su
# modelo: spec-kit no le dice al ejecutable qué paso es, y los dos jueces de la
# revisión final corren a la vez en el mismo fan-out. Se quita antes de llamar a
# Claude Code y se apunta con la sesión (abajo):
#
#   --model claude-opus-5-5@xhigh#juez_spec:modelo_juez  →  --model claude-opus-5-5 --effort xhigh
#
# Es además la única puerta por la que el workflow y scripts/paso.sh abren una
# sesión con modelo, así que aquí se pone lo que esa sesión no puede hacer
# (ADR 0032; al final del fichero).
#
# El valor llega a argv desde los inputs del workflow: se valida con una
# expresión estricta y cualquier otro formato aborta con exit 2. El prompt
# (argumento de -p) se copia sin interpretarlo.
set -euo pipefail

patron='^(fable|opus|sonnet|haiku|claude-[a-z0-9-]+)(@(low|medium|high|xhigh|max))?(#([a-z0-9_]+):(modelo_[a-z]+))?$'

args=()
ajustes_de_quien_llama=""
modelo=""; esfuerzo=""; paso=""; rol=""; sesion=""; reanuda=""
while [ $# -gt 0 ]; do
  case "$1" in
    -p|--print)
      args+=("$1"); shift
      if [ $# -gt 0 ]; then args+=("$1"); shift; fi
      ;;
    --model)
      [ $# -ge 2 ] || { echo "claude-modelo: --model sin valor" >&2; exit 2; }
      if [[ ! "$2" =~ $patron ]]; then
        echo "claude-modelo: modelo inválido '$2' (formato: <modelo>[@low|medium|high|xhigh|max][#<paso>:<rol>])" >&2
        exit 2
      fi
      modelo="${BASH_REMATCH[1]}"; esfuerzo="${BASH_REMATCH[3]}"; paso="${BASH_REMATCH[5]}"; rol="${BASH_REMATCH[6]}"
      args+=(--model "$modelo")
      if [ -n "$esfuerzo" ]; then args+=(--effort "$esfuerzo"); fi
      shift 2
      ;;
    --session-id)
      [ $# -ge 2 ] || { echo "claude-modelo: --session-id sin valor" >&2; exit 2; }
      sesion="$2"; args+=("$1" "$2"); shift 2
      ;;
    -r|--resume|-c|--continue)
      reanuda=1; args+=("$1"); shift
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
#   · Ni plugins ni servidores MCP: los plugins que los ajustes de usuario o los del
#     proyecto tienen activados llegan apagados (`enabledPlugins`), y `--strict-mcp-config`
#     deja fuera los servidores de .mcp.json y de la configuración del usuario. Las skills,
#     las herramientas y las instrucciones de cada uno entran en el contexto de la sesión
#     y se vuelven a pagar en cada turno que le queda (una sesión de H25 llevaba 60–70 k
#     tokens fijos antes de leer nada), y ningún paso usa ninguno: ni el `kitlegal@synced`
#     ni el MCP `gsc` de la web tienen nada que hacer en un paso.
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
# Los plugins activados en los dos ajustes que la sesión lee (`--setting-sources user,project`):
# los del usuario —el directorio de Claude Code, que CLAUDE_CONFIG_DIR puede cambiar— y los
# del proyecto. Cada uno va a los ajustes del paso con `false`, que gana a ambos. Las skills
# de la cuenta de claude.ai (anthropic-skills:*) no son plugins: las apaga
# `syncClaudeAiSkills: false`, que en --settings las bloquea y las oculta solo en esta
# invocación, sin mover las ya sincronizadas.
de_ajustes=()
for f in "$raiz/.claude/settings.json" "${CLAUDE_CONFIG_DIR:-$HOME/.claude}/settings.json"; do
  if [ -f "$f" ]; then de_ajustes+=("$f"); fi
done
plugins='{}'
if [ ${#de_ajustes[@]} -gt 0 ]; then
  plugins="$(jq -cn '[inputs | (.enabledPlugins // {}) | to_entries[] | select(.value == true) | {(.key): false}] | add // {}' \
    "${de_ajustes[@]}")" \
    || { echo "claude-modelo: no se han podido leer los plugins activados en los ajustes de usuario y de proyecto" >&2; exit 2; }
fi
ajustes="$(jq -cn --argjson base "$base" --argjson plugins "$plugins" --arg home "$HOME" --arg secretos "${KITLEGAL_SECRETOS:-}" \
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
      permissions: {deny: ($vedados | map("Read(/" + . + "/**)"))},
      enabledPlugins: $plugins,
      syncClaudeAiSkills: false
    } as $paso
  | ($base * $paso)
  | .permissions.deny = (($base.permissions.deny // []) + $paso.permissions.deny)')" \
  || { echo "claude-modelo: no se han podido componer los ajustes de la sesión (¿--settings con un JSON que no vale?)" >&2; exit 2; }

# El modelo con el que corre de verdad cada sesión (ADR 0033; scripts/workflow/modelos.sh).
# La sesión lleva un id que se le fija aquí, y con él se apunta en gates/modelos.jsonl del
# hito antes de abrirla; el modelo real lo saca después scripts/workflow/modelos.sh resolver
# del transcript que Claude Code escribe con ese id. Se registra antes del `exec` porque el
# wrapper tiene que terminar en él: spec-kit mata con SIGKILL lo que lanzó cuando vence el
# plazo de un paso, y una sesión hija de este proceso quedaría viva. El registro no cambia ni
# la salida ni el código de salida de la sesión: si no se puede escribir, se dice en la
# salida de error y la sesión se abre igual.
if [ -n "$modelo" ]; then
  if [ -z "$sesion" ] && [ -z "$reanuda" ]; then
    sesion=$(od -An -N16 -tx1 /dev/urandom | tr -d ' \n' \
      | sed -E 's/^(.{8})(.{4}).(.{3})(.)(.{3})(.{12})$/\1-\2-4\3-\4\5-\6/' \
      | awk '{ v = substr($0, 20, 1); sub(/^./, substr("89ab", index("0123456789abcdef", v) % 4 + 1, 1), v)
               print substr($0, 1, 19) v substr($0, 21) }')
    args+=(--session-id "$sesion")
  fi
  if ! error=$("$raiz/scripts/workflow/modelos.sh" registrar "$sesion" "$paso" "$rol" "$modelo" "$esfuerzo" 2>&1 >/dev/null); then
    echo "claude-modelo: no se ha podido registrar la sesión en gates/modelos.jsonl (${error:-sin detalle}); la sesión se abre igual" >&2
  fi
fi

export KITLEGAL_PASO_DE_WORKFLOW=1
export CLAUDE_CODE_DISABLE_BACKGROUND_TASKS=1
export PATH="$raiz/scripts/workflow/sin-modelo:$PATH"

exec "$claude" ${args[@]+"${args[@]}"} --strict-mcp-config --setting-sources user,project --settings "$ajustes"
