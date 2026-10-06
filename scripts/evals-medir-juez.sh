#!/usr/bin/env bash
# Ejecuta la medida del juez con modelo de las evals de una skill e imprime la medida (FR-050 a FR-054 de H24;
# contracts/medida-del-juez.md §7 y contracts/job-de-evals.md §3 de H24; ADR 0037). Comprueba su argumento y sus
# variables antes de ejecutar nada; crea su directorio temporal y lo borra al terminar, con el código que sea; y ejecuta
# con una sola orden de Go TestMedidaDelJuez, que lee el juez de la skill, reconstruye sus casos etiquetados sin abrir
# ninguna sesión de evals, vota cada uno con scripts/evals-voto.sh y el Claude Code del juez, y deja en medida.json del
# temporal la medida de lo que hay: las huellas de la rúbrica y de los casos, el modelo y la versión recibidos y sus dos
# recuentos. No comprueba la medida versionada ni dice si corresponde: en esa propuesta lo dice make ci. No usa nada
# posterior a bash 3.2, el de macOS.
#
#   make evals-medir-juez SKILL=<skill>
#
# Abre sesiones con modelo y consume la credencial de Claude Code: la lanza una persona, en el job de evals, con la
# etiqueta evals-medir-juez en una propuesta de cambio o con la entrada medir_al_juez del flujo lanzado a mano; ni make
# ci, ni los ganchos, ni una ejecución normal del job, ni ninguna tarea del workflow (FR-050, FR-093). No escribe nada en
# el repositorio: la medida impresa la versiona esa persona, si se cumple (FR-054).
#
# Variables obligatorias: MODELO_DEL_JUEZ, el id del modelo del juez, y VERSION_DE_CLAUDE_CODE_DEL_JUEZ, la versión de
# Claude Code de sus votos, que van a la medida tal cual; CONCURRENCIA_DE_EVALS, cuántos casos se votan a la vez como
# mucho; COMMIT_EVALUADO, que va en el origen de la medida; CLAUDE_DEL_JUEZ, la ruta absoluta del claude de esa versión,
# un fichero ejecutable; y CLAUDE_CODE_OAUTH_TOKEN, que lee Claude Code.
set -euo pipefail
cd "$(dirname "$0")/.."

# La skill, sola y con valor: make la da vacía si no recibe SKILL.
if [[ $# -ne 1 ]] || [[ -z "$1" ]]; then
	echo "evals-medir-juez: uso: scripts/evals-medir-juez.sh <skill>" >&2
	exit 1
fi

# Las variables, antes de construir nada y de pedir ningún voto. Que tengan un valor que sirva lo dice TestMedidaDelJuez
# con su error: aquí solo se mira que estén, y que el claude del juez, que falta también si su ruta no es la de un
# fichero ejecutable, se pueda ejecutar.
for variable in MODELO_DEL_JUEZ VERSION_DE_CLAUDE_CODE_DEL_JUEZ CONCURRENCIA_DE_EVALS COMMIT_EVALUADO CLAUDE_DEL_JUEZ \
	CLAUDE_CODE_OAUTH_TOKEN; do
	if [[ -z "${!variable:-}" ]]; then
		echo "evals-medir-juez: falta $variable" >&2
		exit 1
	fi
done
if [[ ! -f "$CLAUDE_DEL_JUEZ" ]] || [[ ! -x "$CLAUDE_DEL_JUEZ" ]]; then
	echo "evals-medir-juez: falta CLAUDE_DEL_JUEZ" >&2
	exit 1
fi

# El temporal, en TMPDIR o, sin él, en /tmp, el directorio temporal por defecto de POSIX, se borra al salir, también
# cuando el guion termina con otro código: todo lo que la medida escribe va dentro —la medida y, en su tmp/, que es el
# TMPDIR de go test, el directorio de trabajo de go, la reconstrucción de los casos y el directorio del juez, con el
# HOME y la configuración de Claude Code de sus votos—. Fuera quedan solo las cachés de Go. Su ruta es absoluta:
# TestMedidaDelJuez la recibe por bandera y go test lo ejecuta en el directorio de su paquete.
temporal="$(mktemp -d "${TMPDIR:-/tmp}/kitlegal-medida-del-juez.XXXXXX")"
trap 'rm -rf -- "$temporal"' EXIT
temporal="$(cd "$temporal" && pwd -P)"
mkdir "$temporal/tmp"

# Sin límite de tiempo de go test: el de cada voto es su tope, el de la ejecución es el del trabajo que la lanza, y
# SIGINT y SIGTERM cortan los votos abiertos. Las dos salidas de go test son las del guion: si un defecto no queda
# marcado o un correcto queda marcado, o si un caso queda sin juzgar, el error de TestMedidaDelJuez nombra cada caso.
codigo=0
TMPDIR="$temporal/tmp" go test -tags evals -count=1 -timeout 0 -run '^TestMedidaDelJuez$' ./internal/evals/ -args \
	-skill "$1" -modelo-del-juez "$MODELO_DEL_JUEZ" -version-del-juez "$VERSION_DE_CLAUDE_CODE_DEL_JUEZ" \
	-claude-del-juez "$CLAUDE_DEL_JUEZ" -concurrencia "$CONCURRENCIA_DE_EVALS" -commit "$COMMIT_EVALUADO" \
	-salida "$temporal/medida.json" || codigo=$?

# La medida entera entre sus dos marcas, que es de donde la saca del registro quien la lanzó; termina en salto de línea,
# así que cada marca queda en su propia línea. Se imprime también cuando no se cumple, con sus recuentos. Si
# TestMedidaDelJuez no la escribió —un caso sin juzgar, o un error anterior a los votos—, no se imprime ninguna marca:
# una medida con casos sin juzgar no es una medida (FR-053).
if [[ -f "$temporal/medida.json" ]]; then
	echo "--- inicio de medida.json ---"
	cat "$temporal/medida.json"
	echo "--- fin de medida.json ---"
fi

exit "$codigo"
