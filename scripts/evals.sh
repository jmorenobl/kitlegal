#!/usr/bin/env bash
# Ejecuta las evals de una skill con Claude Code y escribe su informe (FR-070, FR-071; contracts/job-de-evals.md §3 de
# H5; contracts/ejecucion-del-job.md §1 de H7.3). Antes de la primera sesión comprueba todo lo que la evaluación
# necesita: Linux con strace y claude; que no hay Python accesible; que el proxy de las sesiones rechaza; que los
# ficheros de eval están bien formados y que lo grabado sirve sin red cada consulta que necesitan; y que la skill está
# instalada y kitlegal en el PATH. Después, una sola orden de Go, TestEjecucionDelJob, abre las sesiones que pide el
# plan —cada eval con el modelo que decide y con cada modelo informativo, repetida REPETICIONES_DE_EVALS veces—, como
# mucho CONCURRENCIA_DE_EVALS a la vez, cada una con scripts/evals-sesion.sh, preparada justo antes, en su propio
# directorio y con su tope de 240 s, con la skill tal como la deja make install, sin red de ninguna fuente y bajo
# strace; no abre ninguna más tras el mensaje del límite de uso de la cuenta; mide su duración; y las juzga todas en el
# informe, con los umbrales que deciden (FR-030, FR-031, FR-044, FR-050 y FR-051 de H7.3).
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
# pasar (ADR 0016); CONCURRENCIA_DE_EVALS, cuántas sesiones se abren a la vez como mucho; y COMMIT_EVALUADO.
# Opcionales: OBJETIVO_DE_DURACION_DE_EVALS, los segundos que el job admite para sus sesiones, 0 si no está, que es no
# tener objetivo; PRUEBA_DE_RED, que con el valor true añade la sesión de prueba de red de la primera eval (§6);
# CLAUDE_CODE_OAUTH_TOKEN, que lee Claude Code; y RUNNER_TEMP o TMPDIR, donde vive la carpeta de salida
# kitlegal-evals-<skill>.
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

# La concurrencia, también obligatoria, es un entero con al menos una sesión a la vez (FR-030 de H7.3): sin ella, o con
# un valor que no lo es, el repartidor no abriría ninguna.
if [[ ! "${CONCURRENCIA_DE_EVALS:-}" =~ $forma_de_entero ]]; then
	echo "evals: la concurrencia ${CONCURRENCIA_DE_EVALS:-} tiene que ser un entero mayor o igual que 1" >&2
	exit 1
fi

# La carpeta de salida se vacía al empezar, de modo que ningún fichero de una ejecución anterior —un sin-python.txt, un
# informe— pase por uno de esta. Su ruta es absoluta: TestEjecucionDelJob la recibe por bandera y go test lo ejecuta en
# el directorio de su paquete, y cada sesión cambia al suyo. Sin RUNNER_TEMP ni TMPDIR, /tmp, el directorio temporal
# por defecto de POSIX.
temporal="${RUNNER_TEMP:-${TMPDIR:-/tmp}}"
rm -rf -- "$temporal/kitlegal-evals-$skill"
mkdir -p "$temporal/kitlegal-evals-$skill/sesiones"
salida="$(cd "$temporal/kitlegal-evals-$skill" && pwd -P)"

# 2. Linux, que es donde hay strace, y las dos órdenes de la sesión. El tope de cada sesión lo pone el repartidor, en
# Go, sin timeout (research.md D9 de H7.3).
if [[ "$(uname -s)" != Linux ]] || ! command -v strace >/dev/null || ! command -v claude >/dev/null; then
	echo "evals: necesita Linux con strace y claude" >&2
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

# 6. La skill instalada con make install, que es la única que ve la sesión (FR-077), y kitlegal en el PATH, que es desde
# donde la skill lo invoca (FR-127 de H19; contracts/skills-e-invocacion.md §6).
if [[ ! -f "$HOME/.claude/skills/$skill/SKILL.md" ]]; then
	echo "evals: la skill $skill no está instalada" >&2
	exit 1
fi
if ! command -v kitlegal >/dev/null; then
	echo "evals: kitlegal no está en el PATH" >&2
	exit 1
fi

# Sesiones e informe (contracts/ejecucion-del-job.md §1, §3 y §5 de H7.3): TestEjecucionDelJob compone el plan, reparte
# sus sesiones y escribe informe.md e informe.json con la duración de las sesiones. Cada sesión se prepara justo antes
# de abrirla, porque buscar y metadatos caducan a los 300 s, y ninguna se reintenta; el informe no deja pasar la que no
# terminó. La orden falla con un error —una falta en la preparación, una interrupción con SIGINT o SIGTERM—, sin
# escribir el informe, o con el veredicto fallo: por una serie que no pasa, por un umbral que decide y no se cumple, por
# la duración o por sesiones sin medir (FR-037 y FR-043 de H7.3). Sin límite de tiempo de go test: el de la ejecución
# es el tope de la definición del job.
codigo_del_informe=0
go test -tags evals -count=1 -timeout 0 -run '^TestEjecucionDelJob$' ./internal/evals/ -args \
	-skill "$skill" -modelo-que-decide "$MODELO_DE_EVALS" -modelos-informativos "$MODELOS_INFORMATIVOS_DE_EVALS" \
	-repeticiones "$REPETICIONES_DE_EVALS" -umbral "$UMBRAL_DE_EVALS" -concurrencia "$CONCURRENCIA_DE_EVALS" \
	-prueba-de-red="${PRUEBA_DE_RED:-false}" -objetivo-de-duracion "${OBJETIVO_DE_DURACION_DE_EVALS:-0}" \
	-skills "$HOME/.claude/skills" -commit "$COMMIT_EVALUADO" -sin-python "$salida/sin-python.txt" \
	-sesiones "$salida/sesiones" -informe "$salida" ||
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
