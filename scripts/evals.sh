#!/usr/bin/env bash
# Ejecuta las evals de una skill con Claude Code y escribe su informe (FR-070, FR-071; contracts/job-de-evals.md §3 de
# H5). Antes de la primera sesión comprueba todo lo que la evaluación necesita: Linux con strace, claude y timeout; que
# no hay Python accesible; que el proxy de las sesiones rechaza; que los ficheros de eval están bien formados y que lo
# grabado sirve sin red cada consulta que necesitan; y que la skill está instalada. Después abre las sesiones que pide
# el plan —cada eval con el modelo que decide y con cada modelo informativo, repetida REPETICIONES_DE_EVALS veces—, con
# la skill tal como la deja make install, sin red de ninguna fuente y bajo strace, y las juzga todas en el informe.
#
#   make evals SKILL=<skill>
#
# Abre sesiones con modelo y consume la credencial de Claude Code: lo ejecuta el job de evals
# (.github/workflows/evals.yml); ni make ci, ni los ganchos, ni ninguna tarea del workflow. Necesita root o sudo sin
# contraseña para buscar Python en todo el sistema de ficheros.
#
# Variables obligatorias, todas fijadas en la definición del job: MODELO_DE_EVALS, el modelo que decide el veredicto;
# MODELOS_INFORMATIVOS_DE_EVALS, separados por comas, que se ejecutan y se publican como límite inferior sin decidir;
# REPETICIONES_DE_EVALS y UMBRAL_DE_EVALS, las sesiones que se abren de cada eval con cada modelo y cuántas tienen que
# pasar (ADR 0016); y COMMIT_EVALUADO. Opcionales: PRUEBA_DE_RED, que con el valor true añade la sesión de prueba de red
# de la primera eval (§6); CLAUDE_CODE_OAUTH_TOKEN, que lee Claude Code; y RUNNER_TEMP o TMPDIR, donde vive la carpeta
# de salida kitlegal-evals-<skill>.
set -euo pipefail
cd "$(dirname "$0")/.."

# Comprobaciones previas (§3.1): todas antes de la primera sesión, y cualquier fallo termina con código 1.

# 1. El argumento y las variables obligatorias. El argumento tiene la forma del name de una skill (data-model §1.1):
# de él sale la carpeta de salida que el guion vacía, y un nombre con / o .. la sacaría de la carpeta temporal.
forma_de_skill='^[a-z0-9]+(-[a-z0-9]+)*$'
if [[ $# -ne 1 ]] || [[ ! "$1" =~ $forma_de_skill ]]; then
	echo "evals: uso: scripts/evals.sh <skill>" >&2
	exit 1
fi

skill="$1"

for variable in MODELO_DE_EVALS MODELOS_INFORMATIVOS_DE_EVALS REPETICIONES_DE_EVALS UMBRAL_DE_EVALS COMMIT_EVALUADO; do
	if [[ -z "${!variable:-}" ]]; then
		echo "evals: falta $variable" >&2
		exit 1
	fi
done

# Las repeticiones y el umbral son enteros, con al menos una repetición y un umbral que cabe en ellas: un umbral mayor
# que las repeticiones no lo alcanzaría nunca ninguna serie, y uno menor que 1 lo alcanzarían todas (ADR 0016).
forma_de_entero='^[1-9][0-9]*$'
if [[ ! "$REPETICIONES_DE_EVALS" =~ $forma_de_entero ]] || [[ ! "$UMBRAL_DE_EVALS" =~ $forma_de_entero ]] ||
	[[ "$UMBRAL_DE_EVALS" -gt "$REPETICIONES_DE_EVALS" ]]; then
	echo "evals: el umbral $UMBRAL_DE_EVALS tiene que ser un entero entre 1 y las repeticiones $REPETICIONES_DE_EVALS" >&2
	exit 1
fi

# La carpeta de salida se vacía al empezar, de modo que ningún fichero de una ejecución anterior —un sin-python.txt, un
# informe— pase por uno de esta. Su ruta es absoluta: los tests la reciben por bandera y go test los ejecuta en el
# directorio de su paquete, y cada sesión cambia al suyo. Sin RUNNER_TEMP ni TMPDIR, /tmp, el directorio temporal por
# defecto de POSIX.
temporal="${RUNNER_TEMP:-${TMPDIR:-/tmp}}"
rm -rf -- "$temporal/kitlegal-evals-$skill"
mkdir -p "$temporal/kitlegal-evals-$skill/sesiones"
salida="$(cd "$temporal/kitlegal-evals-$skill" && pwd -P)"

# 2. Linux, que es donde hay strace, y las tres órdenes de la sesión.
if [[ "$(uname -s)" != Linux ]] || ! command -v strace >/dev/null || ! command -v claude >/dev/null ||
	! command -v timeout >/dev/null; then
	echo "evals: necesita Linux con strace, claude y timeout" >&2
	exit 1
fi

# 3. Sin Python (FR-073, FR-081; research.md D17 y V56): la búsqueda del paso «Retirar Python del runner» de
# .github/workflows/evals.yml, el mismo arreglo carácter a carácter, como root en todo el sistema de ficheros salvo /proc
# y /sys. Escribe sin-python.txt con la orden, el usuario y el resultado, también cuando encuentra algo; si no puede
# buscar como root o find falla, termina sin escribirlo, porque una búsqueda incompleta no puede decir «ninguno». La
# salida de error de find y de sudo va a la del guion.
if [ "$(id -u)" -eq 0 ]; then como_root=(); else como_root=(sudo -n); fi
busqueda=(find / '(' -path /proc -o -path /sys ')' -prune -o '(' '(' -type f -perm /111 '(' -iname 'python*' -o -iname 'pypy*' ')' ')' -o '(' -type l '(' -iname 'python*' -o -iname 'pypy*' ')' ')' -o '(' '(' -type f -o -type l ')' '(' -iname 'libpython*' -o -iname 'libpypy*' ')' ')' ')' -print)
if ! usuario=$("${como_root[@]}" id -un) || ! encontrado=$("${como_root[@]}" "${busqueda[@]}"); then
	echo "evals: no se pudo buscar Python como root en todo el sistema de ficheros" >&2
	exit 1
fi
{
	printf 'búsqueda: %s\n' "${busqueda[*]}"
	printf 'usuario: %s\n' "$usuario"
	if [ -n "$encontrado" ]; then printf 'resultado:\n%s\n' "$encontrado"; else printf 'resultado: ninguno\n'; fi
} > "$salida/sin-python.txt"
if [ -n "$encontrado" ]; then
	while IFS= read -r ruta; do echo "evals: hay Python accesible: $ruta" >&2; done <<< "$encontrado"
	exit 1
fi

# 4. El puerto al que apunta el proxy de las sesiones está cerrado (FR-074, FR-076; research.md D16). La subshell en la
# condición del if deja el puerto cerrado, el caso normal, en una condición falsa; la orden suelta, con set -e,
# terminaría el guion (research.md V50).
if (exec 3<>/dev/tcp/127.0.0.1/9) 2>/dev/null; then echo "evals: 127.0.0.1:9 acepta conexiones; el proxy de la sesión no bloquearía" >&2; exit 1; fi

# 5. Los ficheros de eval y lo grabado (FR-075; US4-4, US4-6): TestEvalsDelRepositorio lee cada fichero de eval, aplica
# las reglas del conjunto y, con Preparar y ComprobarSinRed, comprueba que lo grabado sirve sin red cada consulta que
# las evals necesitan; TestIdentificadoresDeLasNormas, que los identificadores de la tabla de normas son los grabados.
# Su salida nombra cada fichero mal formado y cada falta.
if ! go test -count=1 -run '^(TestEvalsDelRepositorio|TestIdentificadoresDeLasNormas)$' ./internal/evals/; then
	exit 1
fi

# 6. La skill instalada con make install, que es la única que ve la sesión (FR-077).
if [[ ! -f "$HOME/.claude/skills/$skill/SKILL.md" ]]; then
	echo "evals: la skill $skill no está instalada" >&2
	exit 1
fi

# Sesiones (§3.2). sesion <nombre> <fichero de eval> <modelo> [-prueba-de-red] prepara el directorio de la sesión y la
# ejecuta.
# La preparación y la sesión no se reintentan. Una falta en la preparación termina el guion con código 1; un código de la
# sesión distinto de 0 no lo detiene, pero se escribe siempre, también 0, y el informe no deja pasar la sesión que no
# terminó (§4).
sesion() {
	local nombre="$1" fichero="$2" modelo="$3"
	shift 3

	local d="$salida/sesiones/$nombre"
	mkdir -p "$d/trabajo" "$d/cache" "$d/traza"

	# Preparación (research.md D14): llena cache/ con las consultas necesarias de todas las evals de la skill, justo
	# antes de la sesión porque buscar y metadatos caducan a los 300 s, y escribe pregunta.txt, eval.txt y modelo.txt.
	if ! go test -tags evals -count=1 -run '^TestPrepararSesion$' ./internal/evals/ -args -skill "$skill" -eval "$fichero" -modelo "$modelo" -sesion "$d" "$@"; then
		echo "evals: no se pudo preparar la sesión $nombre" >&2
		exit 1
	fi

	# Sesión: la caché preparada, un proxy que rechaza toda petición salvo la del modelo, sin tráfico no esencial, con el
	# tope de 240 s y bajo strace, en un directorio de trabajo vacío fuera del repositorio y solo con la configuración y
	# las skills de la cuenta (tabla del §3.2; research.md D12, D13 y D16). La traza lleva entera cada cadena de un execve:
	# -s 131072 es el tamaño máximo de un argumento en Linux, y con -s 4096 la instantánea de shell que Claude Code crea
	# antes de la primera orden de Bash sale cortada y deja ilegible la traza (research.md V61 y V62). Sin el aislamiento
	# de subprocesos de Claude Code, que en Linux ejecuta cada orden dentro de bwrap, con un espacio de nombres de PID que
	# deja la traza sin atribuir y el disco de solo lectura; Claude Code sigue sin pasar la credencial del modelo al entorno
	# de las órdenes (research.md V12, V61 y D13).
	local codigo=0
	(
		cd "$d/trabajo"
		env \
			KITLEGAL_CACHE_DIR="$d/cache" \
			HTTP_PROXY=http://127.0.0.1:9 HTTPS_PROXY=http://127.0.0.1:9 \
			http_proxy=http://127.0.0.1:9 https_proxy=http://127.0.0.1:9 \
			NO_PROXY=api.anthropic.com no_proxy=api.anthropic.com \
			CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1 \
			CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0 \
			timeout --kill-after=10s 240s \
			strace -ff -e trace=execve,connect,clone,clone3,fork,vfork -s 131072 -o "$d/traza/t" -- \
			claude -p "$(cat "$d/pregunta.txt")" \
			--model "$modelo" \
			--output-format stream-json --verbose \
			--max-turns 30 \
			--no-session-persistence \
			--setting-sources user \
			--settings '{"sandbox":{"enabled":false}}' \
			--permission-mode bypassPermissions \
			--disallowedTools WebFetch WebSearch \
			> "$d/sesion.jsonl" 2> "$d/sesion.err"
	) || codigo=$?
	printf '%s\n' "$codigo" > "$d/codigo-de-la-sesion"
}

# El plan de sesiones lo calcula Go y lo escribe en plan.tsv: una línea por sesión con su nombre, su fichero de eval, su
# modelo y si lleva el texto de la prueba de red, separados por tabuladores (§3.2). El guion solo lo ejecuta; el informe
# vuelve a calcular las mismas series y exige que estén todas, así que el plan no se escribe dos veces (ADR 0016).
if ! go test -tags evals -count=1 -run '^TestPlanDeSesiones$' ./internal/evals/ -args \
	-skill "$skill" -modelo-que-decide "$MODELO_DE_EVALS" -modelos-informativos "$MODELOS_INFORMATIVOS_DE_EVALS" \
	-repeticiones "$REPETICIONES_DE_EVALS" -prueba-de-red="${PRUEBA_DE_RED:-false}" -plan "$salida/plan.tsv"; then
	echo "evals: no se pudo planificar las sesiones" >&2
	exit 1
fi

# Una sesión por línea del plan, en su orden. La lectura va por un descriptor propio: la sesión lee su pregunta de un
# fichero, pero claude hereda la entrada estándar y se comería el resto del plan.
while IFS=$'\t' read -r nombre fichero modelo prueba_de_red <&3; do
	if [[ "$prueba_de_red" == sí ]]; then
		sesion "$nombre" "$fichero" "$modelo" -prueba-de-red
	else
		sesion "$nombre" "$fichero" "$modelo"
	fi
done 3< "$salida/plan.tsv"

# Informe (§3.3): TestInformeDelJob juzga cada sesión y escribe informe.md e informe.json; falla con el veredicto fallo.
codigo_del_informe=0
go test -tags evals -count=1 -run '^TestInformeDelJob$' ./internal/evals/ -args \
	-skill "$skill" -sesiones "$salida/sesiones" -informe "$salida" \
	-modelo-que-decide "$MODELO_DE_EVALS" -modelos-informativos "$MODELOS_INFORMATIVOS_DE_EVALS" \
	-repeticiones "$REPETICIONES_DE_EVALS" -umbral "$UMBRAL_DE_EVALS" \
	-commit "$COMMIT_EVALUADO" -sin-python "$salida/sin-python.txt" ||
	codigo_del_informe=$?

# La carpeta se vació al empezar: un informe que falta es que el test no lo escribió en esta ejecución.
for f in informe.md informe.json; do
	if [[ ! -f "$salida/$f" ]]; then
		echo "evals: el informe no se escribió: $f" >&2
		exit 1
	fi
done

# Los dos ficheros enteros, cada uno entre sus dos marcas, que es de donde los sacan del registro de la ejecución las
# órdenes de quickstart §12.2 y §12.3 (research.md V48). Los dos terminan en salto de línea, así que cada marca queda en
# su propia línea.
for f in informe.md informe.json; do
	echo "--- inicio de $f ---"
	cat "$salida/$f"
	echo "--- fin de $f ---"
done

if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
	cat "$salida/informe.md" >> "$GITHUB_STEP_SUMMARY"
fi

exit "$codigo_del_informe"
