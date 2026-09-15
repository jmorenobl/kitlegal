# Contrato: el job de evals

FR-070 a FR-077, FR-080 a FR-082, US4-1, US4-8, SC-003, SC-012. Decisiones: research.md D12 (qué se registra de la
sesión), D13 (disparadores, entorno y sesión), D14 (caché por sesión), D16 (garantía de red) y D17 (sin Python). Lo
que no se puede comprobar sin la plataforma está en research.md D22 como supuesto y lo prueba una tarea `[plataforma]`.

## 1. `.github/workflows/evals.yml`

```yaml
# Evals de skills con Claude Code (FR-070). No forma parte de make ci: las evals con modelo cuestan y no son
# deterministas (ADR 0008, «En contra, y asumido»). Contrato: specs/006-h5-skill-boe-legislacion/contracts/job-de-evals.md.
name: evals

on:
  # A mano, sobre la rama que se elija.
  workflow_dispatch:
    inputs:
      prueba_de_red:
        description: "Añade la sesión de prueba de red de la eval 01 (SC-012)"
        type: boolean
        default: false
  # Una vez por semana, sobre la rama principal (research.md D22, S1).
  schedule:
    - cron: '41 4 * * 1'
  # Sobre la rama de un hito antes de fusionar: al poner la etiqueta evals o evals-prueba-de-red en su propuesta de
  # cambio, con el fichero del job de esa rama (FR-070, FR-082; research.md D22, S1).
  pull_request:
    types: [labeled]

jobs:
  evals:
    if: >-
      github.event_name != 'pull_request' ||
      github.event.label.name == 'evals' ||
      github.event.label.name == 'evals-prueba-de-red'
    runs-on: ubuntu-24.04
    timeout-minutes: 120
    permissions:
      contents: read
    env:
      # Un único modelo de gama económica, fijado aquí por su id completo; cambiarlo es un cambio de este fichero (FR-070).
      MODELO_DE_EVALS: claude-haiku-4-5-20251001
      # La versión de Claude Code cuyo comportamiento se comprobó al planificar (research.md, verificación V1-V13).
      VERSION_DE_CLAUDE_CODE: 2.1.270
      SKILL_EVALUADA: boe-legislacion
      COMMIT_EVALUADO: ${{ github.event.pull_request.head.sha || github.sha }}
      PRUEBA_DE_RED: ${{ github.event.label.name == 'evals-prueba-de-red' || inputs.prueba_de_red == true }}
    steps:
      - name: Obtener el código del commit evaluado
        uses: actions/checkout@v7
        with:
          ref: ${{ github.event.pull_request.head.sha || github.sha }}

      - name: Instalar Go y restaurar la caché
        uses: actions/setup-go@v7
        with:
          go-version-file: go.mod
          cache: true
          cache-dependency-path: |
            go.sum
            tools/*/go.sum

      - name: Instalar strace y Claude Code
        shell: bash
        run: |
          # Las opciones las fija el propio paso, como las del paso de retirada de Python (research.md V56).
          set -euo pipefail
          sudo apt-get update
          sudo apt-get install -y strace
          npm install -g "@anthropic-ai/claude-code@${VERSION_DE_CLAUDE_CODE}"
          claude --version

      - name: Instalar kitlegal y las skills como las deja make install
        run: make install

      - name: Retirar Python del runner
        shell: bash
        run: |
          # Las opciones las fija el propio paso, sin depender de con cuáles lo invoque la plataforma: una orden que falla
          # fuera de una condición —la búsqueda o un borrado— lo detiene con su código, y una variable sin valor, también
          # (research.md D17, V56 y V60).
          set -euo pipefail
          # Sin Python accesible para la sesión (FR-073, FR-081; research.md D17). Ninguna orden descarta su salida de error.
          # scripts/evals.sh repite la búsqueda antes de la primera sesión (contrato del job §3.1).
          # No consulta ni purga paquetes: en ubuntu-24.04 los de Python son dependencias de paquetes del sistema y la purga
          # termina con 100 sin retirar nada; la búsqueda encuentra el intérprete y sus bibliotecas, y lo que un paquete
          # deja en disco sin ellos no se ejecuta (research.md D17 y V60).
          # Las marcas se componen al ejecutar: el texto del paso, si el registro lo reproduce, no las contiene.
          marca="retirada de Python"
          echo "--- inicio de la $marca ---"

          # Lo que el job usa después de este paso: la retirada no puede llevárselo (research.md D22, S7).
          usados=("$GITHUB_WORKSPACE" "$HOME/.claude")
          kitlegal=$(readlink -e "$GITHUB_WORKSPACE/bin/instalado/kitlegal")
          usados+=("$kitlegal")
          for orden in bash sudo find rm timeout strace claude node go make git; do
            if ! ruta=$(command -v "$orden"); then echo "falta $orden" >&2; exit 1; fi
            real=$(readlink -e "$ruta")
            usados+=("$ruta" "$real")
          done
          contiene_algo_usado() {
            local u
            for u in "${usados[@]}"; do
              case "$u" in "$1" | "${1%/}"/*) return 0 ;; esac
            done
            return 1
          }

          # 1. Todo el sistema de ficheros salvo /proc y /sys, como root: ejecutables y enlaces python* y pypy*, y bibliotecas
          # libpython* y libpypy*. Se retira cada uno; si está en <prefijo>/bin y <prefijo>/lib tiene un python* o un pypy*,
          # se retira la instalación entera, salvo que contenga algo de lo que el job usa.
          busqueda=(find / '(' -path /proc -o -path /sys ')' -prune -o '(' '(' -type f -perm /111 '(' -iname 'python*' -o -iname 'pypy*' ')' ')' -o '(' -type l '(' -iname 'python*' -o -iname 'pypy*' ')' ')' -o '(' '(' -type f -o -type l ')' '(' -iname 'libpython*' -o -iname 'libpypy*' ')' ')' ')' -print)
          echo "búsqueda: ${busqueda[*]}"
          encontrado=$(sudo "${busqueda[@]}")
          while IFS= read -r ruta; do
            if [ -z "$ruta" ]; then continue; fi
            if ! sudo test -e "$ruta" && ! sudo test -L "$ruta"; then continue; fi
            objetivo=$ruta
            directorio=$(dirname "$ruta")
            if [ "$(basename "$directorio")" = bin ]; then
              prefijo=$(dirname "$directorio")
              lib=""
              if sudo test -d "$prefijo/lib"; then
                lib=$(sudo find -H "$prefijo/lib" -mindepth 1 -maxdepth 1 '(' -iname 'python*' -o -iname 'pypy*' ')' -print -quit)
              fi
              if [ -n "$lib" ] && ! contiene_algo_usado "$prefijo"; then objetivo=$prefijo; fi
            fi
            echo "retirado: $objetivo"
            sudo rm -rf -- "$objetivo"
          done <<< "$encontrado"

          # 2. No queda nada, y la retirada no se ha llevado nada de lo que el job usa.
          queda=$(sudo "${busqueda[@]}")
          if [ -n "$queda" ]; then
            while IFS= read -r ruta; do echo "queda Python: $ruta" >&2; done <<< "$queda"
            exit 1
          fi
          for u in "${usados[@]}"; do
            if ! sudo test -e "$u"; then echo "la retirada se llevó algo que el job usa: $u" >&2; exit 1; fi
          done
          echo "búsqueda tras retirar: ninguno"
          echo "--- fin de la $marca ---"

      - name: Ejecutar las evals
        env:
          CLAUDE_CODE_OAUTH_TOKEN: ${{ secrets.CLAUDE_CODE_OAUTH_TOKEN }}
        run: make evals SKILL="$SKILL_EVALUADA"
```

- Permisos mínimos (`contents: read`); ninguna acción de terceros aplica un control por su cuenta: la evaluación entera
  es `make evals`, igual que `nightly.yml` llama a `make verify-sources`.
- `KITLEGAL_RECORD` no aparece en el fichero (FR-074, constitución, «Reglas del modo desatendido»).
- Las versiones de `actions/checkout` y `actions/setup-go` son las de `ci.yml` y `nightly.yml`.
- Ningún paso descarta un error. Los dos pasos de más de una orden, «Instalar strace y Claude Code» y «Retirar Python
  del runner», fijan sus opciones con `set -euo pipefail` como primera orden: una orden que falla fuera de una condición
  detiene el paso con su código, y una variable sin valor, también. De `shell: bash` el paso solo depende para
  ejecutarse con bash; las opciones que la plataforma añada no cambian nada, porque las fija el propio paso. Comprobado
  con el paso de retirada literal (research.md V60 (5)): la búsqueda que termina con error lo detiene con 1 tras la línea
  `búsqueda:` y sin retirar nada; `GITHUB_WORKSPACE` sin valor, con 1; ejecutado con `sh` (`dash`), con 2 en la propia
  línea `set` y sin hacer nada; y el mismo paso sin esa línea, con la búsqueda fallida, sigue, retira lo que encontró y
  termina con 0, que es lo que la línea evita.
- «Retirar Python del runner» no tira la salida de error de ninguna orden y **no consulta ni purga paquetes**: ni
  `dpkg-query` ni `apt-get` (research.md D17). En `ubuntu-24.04` los paquetes de Python son dependencias de paquetes del
  sistema, y en la prueba de red del intento 1 de T030 (ejecución 34922606273, `gates/prueba-de-red.md`) la purga de los
  110 que elegía el filtro del paso anterior terminó con 100 («pkgProblemResolver::Resolve generated breaks», con
  `shim-signed`, `grub-efi-amd64-signed` y `grub2-common` entre las dependencias rotas) antes de buscar nada; V60 (1) lo
  reproduce con un paquete esencial que depende de uno de Python. Sin purga, qué queda lo decide solo la búsqueda: el
  estado de los paquetes no la cambia, y uno retenido no detiene el paso (V60 (5)).
- **Dónde busca** (research.md D17 y V56): como root, en **todo** el sistema de ficheros salvo `/proc` y `/sys`, en los
  que el núcleo no deja crear ficheros y cuyas entradas cambian mientras se recorren; `/dev` se recorre, porque
  `/dev/shm` admite ejecutables. No hay lista de sitios: un intérprete bajo `/opt`, `/usr/share`, `/home`, en un punto
  de montaje o en cualquier otro directorio se ejecuta igual por su ruta absoluta, y la sesión corre con
  `bypassPermissions`. **Qué busca**: ficheros con algún permiso de ejecución y enlaces cuyo nombre empieza, sin
  distinguir mayúsculas, por `python` o `pypy`, y ficheros y enlaces `libpython*` y `libpypy*` con cualquier modo,
  porque una biblioteca se carga sin permiso de ejecución. Un fichero de datos con ese nombre (`python.vim`, una página
  de manual, un `python.go` de solo lectura de la caché de módulos) no se busca ni se toca.
- **Qué retira**: cada ruta encontrada y, si está en `<prefijo>/bin/` y `<prefijo>/lib/` tiene una entrada `python*` o
  `pypy*` —la biblioteca estándar de una instalación de CPython, de PyPy, de conda o de un entorno virtual—, `<prefijo>`
  entero, salvo que contenga algo de lo que el job usa después del paso: el espacio de trabajo, `~/.claude`, el binario
  al que apunta `bin/instalado/kitlegal` y `bash`, `sudo`, `find`, `rm`, `timeout`, `strace`, `claude`, `node`, `go`,
  `make` y `git`, cada una por su ruta y por su destino. Así `/usr`, `/usr/local` o la instalación que contiene `claude`
  pierden solo el fichero, y el `bin/` de una herramienta que no es de Python con un `python3` suelto, también. Los
  paquetes siguen registrados en dpkg con el estado que tenían y pierden lo que la búsqueda encuentra, también los
  ficheros de su registro con esos nombres —las listas y sumas de los `libpython*`
  (`/var/lib/dpkg/info/libpython3.12t64:<arquitectura>.list` y `.md5sums`) y los guiones de mantenimiento ejecutables de
  los `python*` (`python3.12-minimal.postinst`)—, así que ese registro queda incompleto; ningún paso posterior usa dpkg
  ni apt. Lo que un paquete deja en disco y la búsqueda no encuentra (la biblioteca estándar bajo `/usr/lib/python3.12`,
  `dist-packages`, guiones con `#!/usr/bin/python3`) no se ejecuta sin el intérprete (research.md V60 (1) y (2)).
- **Qué comprueba al final**: repite la búsqueda y termina con 1 si queda alguna ruta (`queda Python: <ruta>`, una línea
  por ruta) o si alguna de las usadas ya no existe (`la retirada se llevó algo que el job usa: <ruta>`); si un borrado
  falla —en un sistema de ficheros de solo lectura, como el de un snap—, el paso termina con el error de `rm` (V60 (4)).
  La comprobación 3 de §3.1 repite la misma búsqueda antes de la primera sesión.
- **Qué deja en el registro**: entre `--- inicio de la retirada de Python ---` y `--- fin de la retirada de Python ---`,
  la búsqueda (`búsqueda: find / ( -path /proc -o -path /sys ) …`), una línea `retirado: <ruta>` por cada fichero o
  instalación retirados y `búsqueda tras retirar: ninguno`. Las marcas se componen al ejecutar, así que no están
  literalmente en el texto del paso, que el registro de la ejecución puede reproducir: la sexta orden de quickstart
  §12.2 solo las encuentra en la salida. La tarea `[plataforma]` lo registra en `gates/prueba-de-red.md`; lo que la
  retirada supone del contenido del runner es el supuesto S7 de research.md.

## 2. `make evals`

```make
## evals: ejecuta las evals de una skill con Claude Code (Linux con strace, como root o con sudo; red solo del modelo; fuera de ci)
evals: check-tools
	scripts/evals.sh "$(SKILL)"
```

Fuera de `ci` (FR-070). Sin `SKILL` el guion falla con su uso.

## 3. `scripts/evals.sh <skill>`

Bash con `set -euo pipefail`. Variables que lee: `MODELO_DE_EVALS`, `COMMIT_EVALUADO` (obligatorias),
`PRUEBA_DE_RED` (`true` o cualquier otra cosa), `CLAUDE_CODE_OAUTH_TOKEN` (la usa Claude Code), `RUNNER_TEMP` o `TMPDIR`
para la carpeta de salida `…/kitlegal-evals-<skill>`, que se vacía al empezar.

### 3.1 Comprobaciones previas (todas antes de la primera sesión; cualquier fallo termina con código 1)

| Orden | Qué comprueba | Mensaje |
|---|---|---|
| 1 | argumento y variables obligatorias | `evals: uso: scripts/evals.sh <skill>` / `evals: falta MODELO_DE_EVALS` |
| 2 | `uname -s` es `Linux`; existen `strace`, `claude` y `timeout` | `evals: necesita Linux con strace, claude y timeout` |
| 3 | **sin Python** (FR-073, FR-081; research.md D17, V56): la búsqueda del paso «Retirar Python del runner» (§1), el mismo arreglo de argumentos, no encuentra nada. Se ejecuta como root —tal cual si el guion corre como root; si no, con `sudo -n`— en todo el sistema de ficheros salvo `/proc` y `/sys`. Escribe `sin-python.txt` con la orden, el usuario y el resultado, también cuando encuentra algo; si no puede buscar como root o `find` falla, termina sin escribirlo, porque una búsqueda incompleta no puede decir «ninguno». Forma literal y contenido, debajo de la tabla | `evals: hay Python accesible: <ruta>`, una línea por ruta encontrada / `evals: no se pudo buscar Python como root en todo el sistema de ficheros` |
| 4 | el puerto `127.0.0.1:9`, al que apunta el proxy de la sesión, está cerrado. Forma literal: `if (exec 3<>/dev/tcp/127.0.0.1/9) 2>/dev/null; then echo "evals: 127.0.0.1:9 acepta conexiones; el proxy de la sesión no bloquearía" >&2; exit 1; fi`. La subshell en la condición del `if` hace que el puerto cerrado —el caso normal— solo deje la condición en falso; la orden `exec 3<>/dev/tcp/127.0.0.1/9` suelta, con `set -euo pipefail`, terminaría el guion con código 1 antes de la primera sesión (research.md V50) | `evals: 127.0.0.1:9 acepta conexiones; el proxy de la sesión no bloquearía` |
| 5 | **ficheros de eval y FR-075**: `go test -count=1 -run '^(TestEvalsDelRepositorio\|TestIdentificadoresDeLasNormas)$' ./internal/evals/`; nombra cada fichero mal formado y cada falta de lo grabado (US4-4, US4-6) | la salida del test |
| 6 | `~/.claude/skills/<skill>/SKILL.md` existe (instalada con `make install`, FR-077) | `evals: la skill <skill> no está instalada` |

Forma literal de la comprobación 3 (research.md V56). `busqueda` es, carácter a carácter, el arreglo del paso de §1:

```bash
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
```

La salida de error de `find` y de `sudo` va a la del guion, sin descartarse. Sin ningún Python, `sin-python.txt` es
exactamente este texto, con un salto de línea al final:

```text
búsqueda: find / ( -path /proc -o -path /sys ) -prune -o ( ( -type f -perm /111 ( -iname python* -o -iname pypy* ) ) -o ( -type l ( -iname python* -o -iname pypy* ) ) -o ( ( -type f -o -type l ) ( -iname libpython* -o -iname libpypy* ) ) ) -print
usuario: root
resultado: ninguno
```

Con alguna ruta encontrada, la tercera línea es `resultado:` y le sigue una ruta por línea, en el orden de `find`; el
guion termina con 1 antes de la primera sesión, así que ese contenido no llega a ningún informe. Por eso, en un informe,
`sin_python` (§5) solo puede ser el texto de arriba.

### 3.2 Una sesión por eval

Para cada fichero de `evals/<skill>/` en orden, y además, si `PRUEBA_DE_RED` es `true`, una sesión
`01-…-prueba-de-red` de la primera eval (§6):

1. `d=<salida>/sesiones/<nombre>`, con `trabajo/`, `cache/` y `traza/` vacíos.
2. **Preparación** (research.md D14): `go test -tags evals -count=1 -run '^TestPrepararSesion$' ./internal/evals/
   -args -skill <skill> -eval <fichero> -sesion <d> [-prueba-de-red]`. `TestPrepararSesion` solo lee sus banderas y
   llama a `PrepararSesion` (firma al final de este apartado) con `Evals` = `evals/<skill>` de la raíz
   (`../../evals/<skill>` desde el directorio del paquete, como `../../schemas` en `internal/app`; research.md V46),
   `Grabaciones` = las de H4 y las de H5 en el orden del contrato de evals §5.1, `Fichero` = `-eval`, `Directorio` =
   `-sesion` y `PruebaDeRed` = `-prueba-de-red`; falla si hay alguna falta o un error. Así llena `d/cache` con las consultas necesarias de **todas**
   las evals de la skill y escribe `d/pregunta.txt` (la pregunta, o la de la prueba de red) y `d/eval.txt` (el nombre del
   fichero). Una falta termina el guion con código 1.
3. **Sesión**, justo después de preparar:

```bash
codigo=0
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
      --model "$MODELO_DE_EVALS" \
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
```

| Pieza | Por qué |
|---|---|
| `KITLEGAL_CACHE_DIR` | la caché preparada; nada en `SKILL.md` la nombra (FR-077) |
| `HTTP(S)_PROXY` a `127.0.0.1:9` | toda petición del binario va al proxy del entorno (`http.DefaultTransport`, que `httpx` clona), que no existe: la conexión se rechaza en el propio equipo y la invocación termina sin texto y con código distinto de 0, lleve o no `--offline` (FR-074, FR-076; research.md D16) |
| `NO_PROXY=api.anthropic.com` | el cliente de la API de Claude Code respeta `HTTPS_PROXY` y `NO_PROXY` (V10): su única red es la del proveedor del modelo |
| `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1` | sin tráfico de telemetría ni actualizaciones, que el proxy bloquearía |
| `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB=0` | sin el aislamiento de subprocesos de Claude Code, que es incompatible con la traza y con el contexto de la skill (research.md V12, V61 y D13). Con `1`, el binario 2.1.270 de Linux exige `bwrap` al arrancar y `socat` en la primera orden, fuerza el modo de permisos `default` (ignora `--permission-mode`) y ejecuta cada orden de Bash dentro de `bwrap` con `--unshare-pid`, `--unshare-user` y `--ro-bind / /`: los `clone` de dentro devuelven los números del espacio de nombres de PID nuevo y no los de los ficheros `t.<n>`, así que la traza no se puede atribuir (§4), y el disco queda de solo lectura salvo `/home`, `/root`, `/tmp`, `/var`, `/opt`, `/run` y `/mnt`, como con el sandbox que D16 rechaza. Con `0`, Claude Code tampoco pasa `CLAUDE_CODE_OAUTH_TOKEN` al entorno de las órdenes ni al de los ganchos, pero la variable sigue en el entorno inicial de `claude` y de los procesos que lo lanzan, legible para una orden del mismo usuario en `/proc/<pid>/environ`: el riesgo que asume plan.md §VII. El valor va explícito para que ningún valor por defecto del binario active el aislamiento |
| `strace -ff …` | registro de cada `execve` (argv y código de salida) y de cada `connect`, de la sesión y de sus descendientes, en un fichero por hilo (research.md D12). Con `-s 131072`, el tamaño máximo de un argumento de `execve` en Linux con páginas de 4 KiB (`MAX_ARG_STRLEN`, 32 páginas, con el nulo final), `strace` escribe entera cada cadena de un `execve`: con `-s 4096`, la instantánea de shell que Claude Code 2.1.270 crea con `bash -c -l` antes de la primera orden de Bash, de unos 7 000 octetos, salía cortada con `"...` y `LeerTrazas` declaraba ilegible la traza de toda sesión que ejecuta Bash (research.md V61 y V62). Un argumento cortado, o una lista de más de 131 072 argumentos, que `strace` corta con `...]`, sigue haciendo ilegible la traza (data-model §9, regla 5): el argv de una invocación cortada no se puede comparar. Sin `-e signal=none`: con ella `strace` no escribe `+++ killed by SIG… +++`, y el fichero de todo proceso que muere por una señal se quedaría sin línea final, con tope o sin él; sin ella añade una línea `--- SIGNOMBRE {…} ---` por señal entregada, que `LeerTrazas` admite y no cuenta (research.md V54; data-model §9, regla 5) |
| `timeout 240s` | por debajo de la vigencia de 300 s de `buscar` y `metadatos` en la caché (data-model §8) |
| `--setting-sources user`, `trabajo/` vacío fuera del repositorio | ni `CLAUDE.md` ni las skills del repositorio entran en la sesión: solo lo que dejó `make install` en `~/.claude/skills` |
| `--settings '{"sandbox":{"enabled":false}}'` | el sandbox no reescribe el proxy de las órdenes (research.md D16) |
| `--permission-mode bypassPermissions` | la sesión sin terminal no puede aprobar órdenes; el runner es desechable y el job solo lee el repositorio. El binario la respeta porque `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB` es `0`: con `1` la sustituye por `default` y deniega toda orden de Bash que no se declare con `--allowedTools` (research.md V12 y V61) |
| `--disallowedTools WebFetch WebSearch` | ninguna herramienta de la sesión consulta la web por su cuenta |
| `codigo=0` … `\|\| codigo=$?` y `codigo-de-la-sesion` | el código de salida de la sesión se escribe **siempre**, también 0: `timeout` da 124 al agotar los 240 s y 137 si, pasados los 10 s de `--kill-after`, tuvo que enviar `KILL`; si la sesión termina antes, da el de `strace`, que es el de `claude` (research.md V52 y V53, comprobados con las versiones de Ubuntu 24.04; en el runner, supuesto S10). Lo lee `LeerSesion` (§4) y decide si la sesión terminó y si la cortó el tope; si el fichero falta, la sesión es ilegible y no pasa, en lugar de tomarse por 0. Con 137, `strace` muere por `KILL` y deja sin línea final los ficheros de los procesos que seguían vivos (research.md V54), lo que `LeerTrazas` admite solo en una sesión cortada (§4) |

La preparación y la sesión no se reintentan. Un código de la sesión distinto de 0 no detiene el guion, pero no se
ignora: la eval se juzga con lo que haya en el transcript y en la traza, y **no pasa** si la sesión no terminó (§4), con
el motivo en el informe. Así una eval de no activación cuya sesión murió por el tope, por un error de la API o sin
mensaje `result` no pasa en vacío por no haber activado nada.

`internal/evals.PrepararSesion(s SesionAPreparar) ([]Falta, error)` (`internal/evals/preparar.go`; sin contexto, como
`Preparar`: contrato de evals §5.1 y research.md V59):

```go
type SesionAPreparar struct {
	Evals       string   // directorio de las evals de la skill: se preparan las consultas de todas
	Grabaciones []string // conjuntos de grabaciones, en el orden en que los copia Preparar
	Fichero     string   // nombre, dentro de Evals, de la eval con la que se juzga la sesión
	Directorio  string   // directorio de la sesión; cache/ ya existe y está vacío
	PruebaDeRed bool     // la pregunta lleva además el texto de la prueba de red (§6)
}
```

1. Lee las evals con `LeerConjunto(s.Evals)` (firma en el contrato de evals §1). Si devuelve un `error`, lo devuelve; si
   `MalFormados` no está vacía, devuelve un `error` que nombra cada fichero mal formado con su error. En los dos casos,
   sin preparar ni escribir nada.
2. Si `s.Fichero` no es el `Fichero` de ninguna de las `Evals`, devuelve un `error` que lo nombra, sin preparar ni
   escribir nada.
3. Ejecuta `Preparar(<s.Directorio>/cache, s.Grabaciones, ConsultasNecesarias(<todas las Evals>))`; si devuelve
   faltas o un error, los devuelve tal cual y no escribe `eval.txt` ni `pregunta.txt`.
4. Escribe `<s.Directorio>/eval.txt` con `s.Fichero` y un salto de línea, y `<s.Directorio>/pregunta.txt` con la
   `pregunta` de esa eval y un salto de línea o, con `s.PruebaDeRed`, con la pregunta, una línea en blanco, el texto
   literal de §6 y un salto de línea.

La fija `TestPrepararDirectorioDeSesion` (§9), sin etiqueta y en `make ci`; `TestPrepararSesion` solo la conecta con el
guion.

### 3.3 Informe

```bash
go test -tags evals -count=1 -run '^TestInformeDelJob$' ./internal/evals/ -args \
  -skill "$skill" -sesiones "$salida/sesiones" -informe "$salida" \
  -modelo "$MODELO_DE_EVALS" -commit "$COMMIT_EVALUADO" -sin-python "$salida/sin-python.txt"
```

`TestInformeDelJob` solo lee sus banderas y llama a `EscribirInforme` con `Skill` = `-skill`, `Evals` = `evals/<skill>`
de la raíz, derivado de `-skill` (`../../evals/<skill>` desde el directorio del paquete, V46), `Sesiones` = `-sesiones`, `Destino` = `-informe`, `Modelo` = `-modelo`, `Commit` = `-commit` y
`SinPython` = `-sin-python`; falla si devuelve un error o si el veredicto es `fallo`.

`internal/evals.EscribirInforme(e InformeAEscribir) (Informe, error)` (`internal/evals/informe.go`):

```go
type InformeAEscribir struct {
	Skill     string // skill evaluada: la activación que se busca y el campo skill del informe
	Evals     string // directorio de las evals con las que se juzga: el job lo deriva de la skill; TestInforme pasa el sintético
	Sesiones  string // directorio con un subdirectorio por sesión (§3.2)
	Destino   string // directorio en el que escribe informe.md e informe.json
	Modelo    string // modelo fijado en el job
	Commit    string // commit evaluado
	SinPython string // ruta de sin-python.txt: su contenido entero, byte a byte, es sin_python del informe (data-model §10.3)
}
```

1. Lee las evals con `LeerConjunto(e.Evals)` (firma en el contrato de evals §1). Su `error` es de los que impiden escribir
   el informe (abajo). Cada `FicheroMalFormado` de `MalFormados` va a `ficheros_mal_formados`, con `fichero` y `error`, y
   las sesiones se juzgan con `Evals`, las bien formadas. No aplica `ComprobarConjuntoDeBoeLegislacion`: las reglas del
   conjunto las aplica antes la comprobación 5 de §3.1 (`TestEvalsDelRepositorio/conjunto`), y para el informe un
   directorio con una o dos evals sintéticas es válido.
2. Por cada subdirectorio de `e.Sesiones`, en orden de nombre, lee siempre `eval.txt` (la eval con la que se juzga: la
   sesión de prueba de red se juzga con la 01), `pregunta.txt` (la pregunta que se hizo, para el informe) y la sesión
   con `LeerSesion`; si `LeerSesion` la leyó, su `traza/` con `LeerTrazas(<sesión>/traza, sesion.Cortada)` (§4). Si
   `LeerSesion` devuelve un error, la traza no se lee: sin el código de la sesión no se sabe si la cortó el tope, y la
   sesión ya no pasa. Solo si todo eso se leyó y `eval.txt` nombra una eval bien formada de `e.Evals`, juzga con
   `Juzgar`. El resultado lleva `sesion` (el nombre del subdirectorio) y `eval` (el nombre que da `eval.txt`, sin su
   salto de línea final, o vacío si falta), y los llevan también las entradas de `fuera_de_lo_grabado` y de `red` que
   salen de él (§5). Cada fichero que falta o no se puede leer —`eval.txt`, `pregunta.txt`, la sesión o la traza— deja
   la sesión sin pasar con un motivo `sesión ilegible: <fichero>: <error>`, uno por fichero y en ese orden, y lo mismo
   un `eval.txt` que no nombra ninguna eval bien formada de `e.Evals`, con un error que da el nombre que lleva: ningún
   fichero del guion que falta se ignora. Una sesión cortada por el tope cuya traza solo tiene lo que deja el corte
   (data-model §9, regla 6) no es ilegible: lleva el motivo de data-model §10.1, y sus invocaciones, sus conexiones y
   sus entradas de `red` y de `fuera_de_lo_grabado` se informan (FR-071, FR-076).
3. Toda eval bien formada de `e.Evals` que ningún `eval.txt` nombra no tiene sesión que la juzgue: da el motivo
   `<fichero>: sin ninguna sesión` en la raíz del informe (p. ej. `11-no-activa-programacion.yaml: sin ninguna sesión`)
   y el veredicto es `fallo`. Si `e.Evals` no tiene ninguna eval bien formada, el motivo es
   `ninguna eval bien formada que juzgar` y el veredicto, `fallo`. Sin estas dos reglas, la sesión que falta de una
   eval, un directorio de sesiones vacío o un directorio de evals vacío darían `aprobado` sin haber evaluado nada de lo
   que había que evaluar (data-model §10.3).
4. Escribe `informe.md` e `informe.json` (§5) en `e.Destino` **antes** de devolver el `Informe` con su veredicto.

El `error` es solo para lo que impide escribir el informe —`e.Evals`, `e.Sesiones` o `e.SinPython` que no se pueden
leer, o un `e.Destino` en el que no se puede escribir— y nombra el directorio o el fichero. Todas las entradas se leen
antes de escribir nada: con cualquiera de esos errores no queda en `e.Destino` ni `informe.md` ni `informe.json`, de modo
que un `sin_python` vacío o a medias no llega nunca a un informe (FR-081). Lo fija `TestEscribirInformeSinSusEntradas`
(§9). Ningún fichero que escribe el
guion se queda sin lector: `sin-python.txt` va al informe; `eval.txt` y `pregunta.txt`, a `EscribirInforme`;
`sesion.jsonl`, `codigo-de-la-sesion` y `sesion.err`, a `LeerSesion`; `traza/`, a `LeerTrazas`, con el corte que da
`LeerSesion`. Todo lo que decide el
veredicto lo fija `TestInforme` en `make ci` (§9), y el test con etiqueta solo lo conecta con el guion.

Después, el guion imprime los dos ficheros en la salida estándar, cada uno entre dos líneas literales y en este orden:

```text
--- inicio de informe.md ---
<informe.md entero>
--- fin de informe.md ---
--- inicio de informe.json ---
<informe.json entero>
--- fin de informe.json ---
```

Si el test no dejó alguno de los dos ficheros, el guion escribe `evals: el informe no se escribió: <fichero>` y termina
con 1; si los dejó, añade `informe.md` a `$GITHUB_STEP_SUMMARY` si esa variable existe y termina con el código del test.
Los dos ficheros terminan en salto de línea (§5), así que cada marca queda en su propia línea. El registro de la
ejecución conserva el informe entero entre marcas, y la tarea `[plataforma]` lo saca de `gh run view --log` con las
órdenes de quickstart §12.2 y §12.3, que imprimen todo lo que hay de cada marca de inicio a su marca de fin, sin tope de
líneas, y fallan sin imprimir nada si falta cualquiera de las cuatro (research.md V48).

## 4. Qué se registra de la sesión (research.md D12)

| Dato | Fuente | Por qué no otra |
|---|---|---|
| activación | `sesion.jsonl`: bloque `tool_use` con `name: "Skill"` e `input.skill` = skill (V6) | es la llamada con la que Claude Code carga una skill |
| respuesta | `sesion.jsonl`: mensaje `result` con `subtype: "success"` e `is_error: false`, campo `result` (V8, V47); en otro caso, vacía | es la respuesta final |
| modelo y versión | `sesion.jsonl`: mensaje `system`/`init`, `model` y `claude_code_version` (V7) | lo declara la propia sesión |
| código de la sesión | `codigo-de-la-sesion`: un entero en una línea (§3.2) | el transcript de una sesión que `timeout` cortó o que falló no dice por qué terminó |
| fin de la sesión | último mensaje de `sesion.jsonl`: su `type` y, si es `result`, su `subtype` e `is_error` (V47), con el texto fijo de data-model §10.1 (`result success`, `result error_max_turns`, `result success con is_error`, `system`, `sin mensajes`) | distingue la sesión que respondió de la cortada por el tope, por un error de la API o por agotar los turnos |
| salida de error | `sesion.err`, entera | dice en el informe por qué no terminó una sesión |
| invocaciones (argv, código) | trazas `traza/t.<n>`, un fichero por hilo (V53): la última línea `execve(…) = 0` de un applet en un proceso, con su argv, y la línea final del fichero de su hilo principal, o sin código si el tope cortó la sesión antes de que la escribiera (V54; data-model §9, regla 6) | el resultado de la herramienta Bash no lleva código de salida (V9) y una orden puede encadenar varias invocaciones |
| conexiones | trazas: los `connect(…)` del hilo principal de la invocación posteriores a esa `execve` y los de los demás hilos de su proceso, atribuidos por `CLONE_THREAD` de forma transitiva, con `clone` y con `clone3` (data-model §9) | observación del sistema, independiente de cómo se bloquea |

**Sesión terminada** = código de la sesión 0 **y** último mensaje del transcript `result` con `subtype: "success"` e
`is_error: false`. Cualquier otra cosa es una sesión sin terminar, con el motivo de data-model §10.1 (`tope de 240 s
agotado (código 124)`, `terminada por señal tras el tope (código 137)`, `código N`, `sin mensaje result`,
`result con subtype <subtype>`, `result con is_error`). Una eval con la sesión sin terminar **no pasa**, sea positiva o
de no activación, y el motivo va a sus `motivos` y a los del informe (§5): si pasara, una eval de no activación daría
verde sin que se hubiera evaluado nada. La traza de una sesión cortada por el tope se lee con lo que admite data-model
§9, regla 6: su motivo es el del tope y no `sesión ilegible`, y lo que hicieron sus invocaciones se informa (§5).

`internal/evals.LeerTrazas(dir string, cortada bool) ([]Invocacion, error)`, que lee los ficheros `t.<pid>` que
`strace -ff` deja en el `traza/` de una sesión, y `LeerSesion(dirDeSesion string) (Sesion, error)`, que lee del
directorio de la sesión `sesion.jsonl`, `codigo-de-la-sesion` y `sesion.err`, y pone `Sesion.Cortada` a verdadero si el
código es 124 o 137. `LeerTrazas` no lee `codigo-de-la-sesion`: el corte lo decide `LeerSesion` y quien llama se lo pasa
(`EscribirInforme`, §3.3), de modo que no hay dos lectores del mismo fichero que puedan discrepar. En `Invocacion`,
`Codigo` es `*int`, `nil` en la invocación sin código (data-model §9). Una traza, un transcript, un código o una salida de error que
faltan o no se pueden leer no se ignoran —el guion escribe siempre los tres ficheros de la sesión, `sesion.err` vacío si
no hubo salida de error—, y un `codigo-de-la-sesion` ausente o que no es un entero no se toma por 0: la eval queda con el
motivo `sesión ilegible: <fichero>: <error>` y no pasa.

**Lectura de la traza** (`LeerTrazas`; reglas completas en data-model §9; formatos comprobados en research.md V53,
V54 y V63, y en el runner, supuesto S4):

- Un hilo creado por una llamada `clone` o `clone3` cuyas banderas incluyen `CLONE_THREAD` pertenece al proceso del hilo
  que lo creó, y la atribución es **transitiva**: el runtime de Go crea sus hilos con `clone`, no con `clone3`, y a menudo
  desde un hilo que no es el principal (research.md V51 y V53). Las conexiones de cualquier hilo del proceso de una
  invocación, posteriores a su `execve`, son de esa invocación. Un hilo creado sin `CLONE_THREAD`, o con `fork` o
  `vfork`, es el hilo principal de un proceso nuevo: Claude Code 2.1.270 de x86_64 crea con `vfork` los procesos de sus
  órdenes.
- `strace` alinea el resultado en su columna 40 (`-a 40`, su valor por defecto): entre el paréntesis de cierre y `= `
  escribe un espacio y, si la llamada es más corta, el relleno de espacios hasta esa columna. `LeerTrazas` admite uno o
  más espacios ahí y nada más: la línea `vfork()` del runner lleva 33 (data-model §9, regla 5; research.md V63). En las
  sondas de V53, V54, V61 y V62 toda llamada del filtro pasaba de 40 columnas y ninguna línea llevaba relleno; una
  llamada con argumentos también puede quedarse corta, como `clone(child_stack=NULL, flags=SIGCHLD)`, de 38 columnas y
  con dos espacios (V63), y se lee igual.
- Un `connect` de un fichero no atribuido a ninguna invocación de applet —el de `claude` a la API del modelo, que es
  pública, o el de `bash`— se ignora.
- Las líneas de señal (`--- SIGNOMBRE {…} ---`) se admiten y no cuentan. La traza se toma sin `-e signal=none` para que
  un proceso que muere por una señal deje `+++ killed by SIG… +++`, y su invocación, un código distinto de 0; a cambio,
  cada señal entregada deja una de esas líneas (research.md V54).
- Lo que no se entiende no se ignora: exactamente un fichero carece de la línea `clone`, `clone3`, `fork` o `vfork` que
  lo crea (el del proceso que arrancó `strace`), cada línea tiene una de las formas de data-model §9 con su resultado y
  cada fichero termina en su línea final. Otra cosa hace la traza ilegible, con un error que nombra el fichero, el número
  de línea y su texto, o los ficheros sin origen, y la eval no pasa. Si un formato distinto en el runner dejara hilos o
  conexiones sin atribuir en silencio, `red` saldría vacío en falso (FR-076).
- Con `cortada` —el tope cortó la sesión—, y solo entonces, se admite lo que puede dejar el corte y nada más
  (data-model §9, regla 6; research.md V54): ficheros sin línea final o vacíos, los de los procesos vivos cuando
  `timeout` mata `strace` con `KILL`, y una llamada con el resultado `? ERRNO (descripción)` a la que en su fichero solo
  siguen líneas de señal y, como mucho, la línea final: la que interrumpió la señal. El origen, las demás
  líneas y la atribución se comprueban igual; la invocación cuyo hilo principal no llegó a su línea final queda sin
  código, y un `connect` sin resultado se clasifica por su dirección. Sin esa excepción, toda sesión cortada saldría como
  `sesión ilegible`, sus conexiones no llegarían a `red` ni sus invocaciones a `fuera_de_lo_grabado` (FR-071, FR-076), y
  un tope en la prueba de red se tomaría por un formato de traza distinto (S4).

## 5. Informe (FR-071)

`informe.json` (clave → contenido; data-model §10):

```json
{
  "skill": "boe-legislacion",
  "modelo": "claude-haiku-4-5-20251001",
  "modelos_de_sesion": ["claude-haiku-4-5-20251001"],
  "versiones_de_claude_code": ["2.1.270"],
  "commit": "<sha de 40 cifras>",
  "sin_python": "<contenido de sin-python.txt>",
  "ficheros_mal_formados": [],
  "veredicto": "aprobado",
  "motivos": [],
  "fuera_de_lo_grabado": [
    {"sesion": "01-lpac-articulo-21-prueba-de-red", "eval": "01-lpac-articulo-21.yaml", "orden": "boe articulo BOE-A-2015-10565 a9998 --json", "codigo": 5},
    {"sesion": "01-lpac-articulo-21-prueba-de-red", "eval": "01-lpac-articulo-21.yaml", "orden": "boe articulo BOE-A-2015-10565 a9998 --offline --json", "codigo": 4}
  ],
  "red": [],
  "evals": [
    {
      "sesion": "01-lpac-articulo-21",
      "eval": "01-lpac-articulo-21.yaml",
      "activa": true,
      "activada": true,
      "comandos_ejecutados": ["bloque boe BOE-A-2015-10565 a21"],
      "comandos_ausentes": [],
      "citas_encontradas": ["BOE-A-2015-10565 a21"],
      "citas_ausentes": [],
      "invocaciones": [{"orden": "boe articulo BOE-A-2015-10565 a21 --json", "codigo": 0, "conexiones": []}],
      "fuera_de_lo_grabado": [],
      "otras_fallidas": [],
      "llegadas_a_la_red": [],
      "respuesta": "…",
      "codigo_de_la_sesion": 0,
      "fin_de_la_sesion": "result success",
      "sesion_terminada": true,
      "motivos": [],
      "pasa": true
    }
  ]
}
```

Es el informe de una ejecución con la prueba de red (§6): `evals` lleva también la sesión
`01-lpac-articulo-21-prueba-de-red`, juzgada con la misma eval que `01-lpac-articulo-21`, y por eso cada entrada de
`fuera_de_lo_grabado` nombra la sesión y la eval. Una entrada de `red` tiene la misma forma con `destino` en lugar de
`codigo`: `{"sesion": "01-lpac-articulo-21", "eval": "01-lpac-articulo-21.yaml", "orden": "boe articulo BOE-A-2015-10565 a9998 --json", "destino": "203.0.113.7:443"}`.
Dentro de cada eval, `fuera_de_lo_grabado`, `otras_fallidas` y `llegadas_a_la_red` llevan solo `orden` y `codigo` (o
`destino`), porque la sesión y la eval son las del resultado.

Cada entrada de `invocaciones` lleva `orden`, `codigo` y `conexiones`: una entrada por pareja distinta de `destino` y
`clase` (`local`, `bloqueada` o `red`), en el orden en que aparece por primera vez (data-model §9 y §10.2). Se listan todas,
también las que no cambian nada, porque son lo que muestra que la traza se leyó y se atribuyó (research.md S4). En la
sesión `01-lpac-articulo-21-prueba-de-red`, las dos invocaciones de `a9998` son
`{"orden": "boe articulo BOE-A-2015-10565 a9998 --json", "codigo": 5, "conexiones": [{"destino": "127.0.0.1:9", "clase": "local"}]}`
y `{"orden": "boe articulo BOE-A-2015-10565 a9998 --offline --json", "codigo": 4, "conexiones": []}`. Una entrada de
`ficheros_mal_formados` lleva `fichero` y `error` (contrato de evals §1):
`{"fichero": "02-sin-pregunta.yaml", "error": "02-sin-pregunta.yaml: …"}`.

Una eval de no activación cuya sesión cortó el tope aparece así, y el veredicto es `fallo` aunque `activa` y `activada`
coincidan:

```json
{
  "sesion": "11-no-activa-programacion",
  "eval": "11-no-activa-programacion.yaml",
  "activa": false,
  "activada": false,
  "codigo_de_la_sesion": 124,
  "fin_de_la_sesion": "system",
  "sesion_terminada": false,
  "motivos": ["la sesión no terminó: tope de 240 s agotado (código 124)"],
  "pasa": false
}
```

con `"motivos": ["11-no-activa-programacion: la sesión no terminó: tope de 240 s agotado (código 124)"]` en la raíz del
informe.

Su traza se lee con `cortada` (§4), así que lo que el corte deja en ella no la hace ilegible. Si el corte deja una
invocación sin terminar, esta aparece en `invocaciones` con `"codigo": null` y sus conexiones, y en la sección de su
sesión de `informe.md` con «sin código (sesión cortada)»; no va a `fuera_de_lo_grabado` ni a `otras_fallidas`, sus
conexiones de clase `red` van a `red`, y las invocaciones que habían terminado antes del corte con 4 o 5 van a
`fuera_de_lo_grabado` (`TestInforme/sesion-cortada-con-invocaciones`, §9).

`informe.md` empieza por la línea literal `# Informe de evals de <skill>` —con esta skill,
`# Informe de evals de boe-legislacion`, que comprueba `TestInforme/aprobado`—, y después: veredicto y motivos; la
cabecera, con estas cuatro líneas literales, que llevan los valores de `informe.json` (data-model §10.3), cada lista
separada por «, » y, si está vacía, `ninguno` en la de modelos y `ninguna` en la de versiones:

```text
Modelo del job: <modelo>
Modelos de las sesiones: <modelos_de_sesion>
Versiones de Claude Code: <versiones_de_claude_code>
Commit: <commit>
```

la sección `## Comprobación sin Python`, con `sin_python` entero en un bloque: la línea ```` ```text ````, el contenido
tal cual, que ya termina en salto de línea, y la línea ```` ``` ````; ficheros mal formados («ninguno» o una
línea por fichero con su error), invocaciones fuera de lo grabado («ninguna» o tabla sesión · eval · orden · código),
peticiones llegadas a la red («ninguna petición llegó a la red de una fuente» o tabla sesión · eval · orden · destino);
una tabla con una fila por sesión (sesión, eval, activa, activada, sesión terminada con su código, comandos ausentes,
citas ausentes, resultado); y una sección por sesión con su eval, la pregunta, las invocaciones con su código (o «sin código (sesión cortada)») y sus
conexiones (el destino y la clase de cada una, o «sin conexiones»), la respuesta y, si la sesión no terminó o no se pudo
leer, su motivo y su salida de error. `informe.md` e `informe.json` terminan en salto de línea (§3.3).

Veredicto: `fallo` si hay ficheros mal formados, si alguna sesión no pasa —también porque no terminó o es ilegible—, si
alguna eval bien formada no tiene ninguna sesión que la juzgue (ningún `eval.txt` la nombra; motivo
`<fichero>: sin ninguna sesión`), si no hay ninguna eval bien formada (motivo `ninguna eval bien formada que juzgar`) o
si alguna invocación tiene una conexión de clase `red`; `aprobado` en otro caso (§3.3, paso 3; data-model §10.3). Los
motivos de la raíz van en el orden de data-model §10.3, y en `informe.md`, en la sección de veredicto y motivos. Las
invocaciones fuera de lo grabado se informan y no cambian ni la eval ni el veredicto (FR-076).

## 6. Prueba de red (SC-012)

Con la etiqueta `evals-prueba-de-red` (o `prueba_de_red` en `workflow_dispatch`), el guion prepara, además de una sesión
por eval, la sesión `01-lpac-articulo-21-prueba-de-red` con `-eval 01-lpac-articulo-21.yaml -prueba-de-red`, y
`PrepararSesion` (§3.2) escribe en su `eval.txt` `01-lpac-articulo-21.yaml` y en su `pregunta.txt` la pregunta de la eval
01, una línea en blanco y este texto literal, que es una sola línea (las comillas invertidas forman parte del texto):

```text
Antes de responder, ejecuta también exactamente estas dos órdenes y di qué devolvieron: `~/.claude/skills/boe-legislacion/scripts/boe articulo BOE-A-2015-10565 a9998 --json` y `~/.claude/skills/boe-legislacion/scripts/boe articulo BOE-A-2015-10565 a9998 --offline --json`.
```

`a9998` no está en ninguna grabación. La sesión se juzga con la eval 01 (FR-072). Lo que hace cada invocación está
comprobado en local con el binario del repositorio, una caché vacía y el mismo proxy que rechaza (research.md V41): sin
`--offline`, la cadena de `httpx` no consigue el `robots.txt` a través del proxy y termina con código 5 (`limite-o-tos`)
sin haber pedido nada a la fuente; con `--offline`, código 4 sin pedir nada. Resultado exigido, que la tarea
`[plataforma]` registra: las dos invocaciones aparecen en `fuera_de_lo_grabado` del informe, cada una con `sesion`
`01-lpac-articulo-21-prueba-de-red` y `eval` `01-lpac-articulo-21.yaml` (en `informe.md`, dos filas de la tabla sesión ·
eval · orden · código), la que va sin `--offline` con código 5 y la que lo lleva con 4; en `invocaciones` de esa sesión
(y en su sección de `informe.md`), la que va sin `--offline` con `conexiones` igual a
`[{"destino": "127.0.0.1:9", "clase": "local"}]` —la del proxy que rechaza: la traza real de research.md V53 tiene tres
`connect` a esa dirección, y el informe da la pareja una vez— y la que lleva `--offline` con `conexiones` vacía; ninguna
en `comandos_ejecutados` de esa sesión; ninguna sesión con el motivo `sesión ilegible`, salvo, si la hubiera, una que
el tope cortó (`codigo_de_la_sesion` 124 o 137); `red` vacío; y `pasa` de esa sesión igual al de la sesión
`01-lpac-articulo-21` si las dos leyeron y citaron `a21`. La evidencia del supuesto S4 con trazas reales del runner
(research.md D12 y D22) es la conexión `local` y que ninguna sesión no cortada (`codigo_de_la_sesion` distinto de 124 y
137, como el 0 de las terminadas) tiene traza ilegible. Una sesión cortada por el tope, legible o no, se anota como
evidencia del supuesto S9 y no del S4: su traza puede quedar sin las líneas finales de los procesos vivos o con una
llamada interrumpida (research.md V54), y no dice nada del formato de una traza completa. El texto de la prueba lo pone el job, no `SKILL.md` (FR-077).

## 7. Ejecución de cierre y aceptación

- **Cierre (FR-082, SC-003).** Con la etiqueta `evals` sobre la propuesta de cambio del hito, después del último cambio
  fuera de `specs/006-h5-skill-boe-legislacion/`. Se registra en `specs/006-h5-skill-boe-legislacion/gates/evals-cierre.md`:
  enlace a la ejecución, `headSha` del commit evaluado, modelo, veredicto, el informe copiado del registro (lo que la
  quinta orden de quickstart §12.3 imprime entre las marcas de §3.3) y, como evidencia de SC-003, la salida entera de la
  última orden de quickstart §12.3, tal cual: `commit evaluado: <sha>`, la lista completa de
  `git diff --name-only <commit evaluado> HEAD` y la línea `todos bajo specs/006-h5-skill-boe-legislacion/`, que la orden
  solo escribe después de comprobar, sin descartar ningún error, que cada línea de la lista empieza por ese directorio
  (si una no, escribe `fuera del directorio del hito: <fichero>` y termina con 1). Lo registrado y lo que imprime la
  orden son lo mismo. Exige 10 de 10 positivas y las dos de no activación en verde, y `red` vacío. Cualquier cambio
  posterior fuera de ese directorio obliga a repetirla: si la etiqueta está puesta se quita, y se vuelve a poner (sin
  quitarla antes no habría evento `labeled` ni ejecución nueva; se quita solo si la propuesta de cambio la lista). La
  ejecución que se registra es la primera de la rama con evento `pull_request` y `workflowName` `evals` (el `name:` de
  §1) creada en el instante del último evento `labeled` de `evals` o después (el más reciente tras ordenar los
  instantes, sin depender del orden de la API), leída por su `databaseId` (quickstart §12.3), nunca la última de la
  lista a ciegas. Ninguna orden pide a `gh` que resuelva el flujo por el nombre de su fichero (`--workflow evals.yml`),
  que mientras tanto solo está en la rama de la propuesta de cambio: filtran las ejecuciones de la rama por el campo
  `workflowName` que expone `gh run list` (research.md V45). Lo que esto supone de la plataforma es el supuesto S12 de
  research.md, que la prueba de red registra antes. «En verde» incluye que la sesión de cada eval terminara (§4). Una sesión cortada por el tope hace fallar la ejecución
  de cierre y se anota en `gates/evals-cierre.md` como evidencia contra el supuesto S9, no del S4 (research.md D22).
- **Aceptación (FR-080, FR-081, SC-001, SC-002).** De esa misma ejecución, en
  `specs/006-h5-skill-boe-legislacion/gates/aceptacion.md`: la respuesta y las invocaciones de las sesiones
  `01-lpac-articulo-21` (norma no fiscal) y `06-irpf-rendimientos-del-trabajo` (fiscal), y, como constancia de cómo se
  comprobó que no había Python (FR-081), el `sin_python` del informe, que es el contenido de `sin-python.txt`: la
  búsqueda como root en todo el sistema de ficheros salvo `/proc` y `/sys` (`búsqueda: find / ( -path /proc -o -path
  /sys ) …`), `usuario: root` y `resultado: ninguno` (§3.1). Con cualquier ruta encontrada, o sin poder buscar como root,
  la comprobación 3 termina con 1 antes de la primera sesión y la ejecución no tiene informe, así que no hay aceptación
  que registrar de ella.

## 8. Prerrequisitos de plataforma (los da de alta una persona)

- Secreto de repositorio `CLAUDE_CODE_OAUTH_TOKEN`: token de larga duración de la suscripción de Claude, generado con `claude setup-token`; el proyecto no usa clave de API de pago por uso, y las sesiones consumen de los límites de la suscripción.
- Etiquetas `evals` y `evals-prueba-de-red`.

La tarea `[plataforma]` los comprueba (`gh secret list`, `gh label list`) y, si falta alguno, se detiene y lo anota.

## 9. Tests que lo fijan

| Test | Qué fija |
|---|---|
| `TestLeerSesion` | un subtest por directorio de `internal/evals/testdata/sesiones/leer-sesion/` (§9.1), con su nombre; el directorio del caso **es** el de una sesión y se pasa tal cual a `LeerSesion`: `activada` (`init` con modelo y versión, `tool_use` `Skill` de la skill, `codigo-de-la-sesion` con 0 y último mensaje `result` con `subtype: success` e `is_error: false`: activada, respuesta, código 0 leído, terminada y no cortada); `no-activada`; `otra-skill-activada` (la activación de otra skill no cuenta); `codigo-distinto-de-cero` (código 1 y `result` `success`: sin terminar y no cortada, motivo `código 1`); `tope-agotado` (código 124 y transcript con solo el mensaje `init`: cortada, motivo `tope de 240 s agotado (código 124)`, que va antes que `sin mensaje result`); `senal-tras-el-tope` (código 137 y transcript con solo el mensaje `init`: cortada, motivo `terminada por señal tras el tope (código 137)`); `sin-result` (código 0 y transcript que termina en un mensaje `assistant`: `sin mensaje result`, respuesta vacía); `error-max-turns` (código 0 y `result` con `subtype: error_max_turns`: `result con subtype error_max_turns`, respuesta vacía); `result-con-is-error` (código 0 y `result` `success` con `is_error: true`: `result con is_error`, respuesta vacía); `sin-fichero-de-codigo` y `codigo-no-entero` (`codigo-de-la-sesion` ausente, o con un texto que no es un entero: error que nombra el fichero, nunca código 0; la eval queda como sesión ilegible y no pasa, `TestInforme/sesion-ilegible`); `sin-transcript` (sin `sesion.jsonl`: error que nombra el fichero); `sin-salida-de-error` (sin `sesion.err`, con transcript y código correctos: error que nombra el fichero); `linea-ilegible` (una línea que no es JSON: error que nombra el fichero y la línea); `sin-mensajes` (`sesion.jsonl` vacío, `codigo-de-la-sesion` con 124 y `sesion.err`: sin error, cortada, respuesta vacía, `Fin` `sin mensajes` y motivo `tope de 240 s agotado (código 124)`: una sesión que el tope corta antes de emitir `init` no es ilegible). En cada subtest que lee el transcript, `Fin` con el texto fijo de data-model §10.1: `result success` en `activada`, `no-activada`, `otra-skill-activada` y `codigo-distinto-de-cero`; `result error_max_turns` en `error-max-turns`; `result success con is_error` en `result-con-is-error`; `assistant` en `sin-result`; `system` en `tope-agotado` y `senal-tras-el-tope`; `sin mensajes` en `sin-mensajes` |
| `TestLeerTrazas` | un subtest por directorio de `internal/evals/testdata/sesiones/leer-trazas/` (§9.1), con su nombre; el directorio del caso **es** el `traza/` de una sesión y se pasa tal cual a `LeerTrazas`, con `cortada` verdadero en `cortada-por-el-tope`, `cortada-con-llamada-interrumpida` y `llamada-interrumpida-antes-del-final` y falso en los demás (reglas de data-model §9; las líneas, con las formas de research.md V53 y, las de señal y las que deja un corte, con las de V54): `argv-escapado` (`execve` de `scripts/boe buscar` con un término escapado en octal y otro en hexadecimal: argv decodificado); `salida-con-codigo` (`+++ exited with 4 +++`: código 4); `muerte-por-senal` (`+++ killed by SIGKILL +++`: código distinto de 0); `hilo-por-clone` (`t.2000`, invocación de `scripts/boe`, crea el hilo 2001 con la línea con la que el runtime de Go crea sus hilos, tal como la imprime `strace` (V51, V53): `clone(child_stack=0x…, flags=CLONE_VM\|CLONE_FS\|CLONE_FILES\|CLONE_SIGHAND\|CLONE_THREAD\|CLONE_SYSVSEM) = 2001`; `t.2001` hace un `connect` a `203.0.113.7:443` con `EINPROGRESS`: atribuido a la invocación de 2000, de clase `red`); `hilo-por-clone3` (`clone3` con `CLONE_THREAD` en `t.2000` que crea el hilo 2001, y el `connect` de `t.2001`, a `127.0.0.1:9` rechazado, atribuido a la invocación de 2000); `hilo-de-un-hilo` (`t.2000` crea 2001 con `clone` como en `hilo-por-clone`; `t.2001`, que no es el hilo principal, crea 2002 con la variante de amd64, `clone(child_stack=0x…, flags=CLONE_VM\|CLONE_FS\|CLONE_FILES\|CLONE_SIGHAND\|CLONE_THREAD\|CLONE_SYSVSEM\|CLONE_SETTLS, tls=0x…) = 2002` (V51: una bandera y un argumento más no cambian nada); el `connect` a `203.0.113.7:443` con `EINPROGRESS` de `t.2002` se atribuye a la invocación de 2000, de clase `red`); `connect-fuera-de-la-invocacion` (`t.1000`, el proceso de `claude`, con su `execve`, un `connect` a `203.0.113.7:443` con `EINPROGRESS` y la línea `clone(child_stack=NULL, flags=CLONE_CHILD_CLEARTID\|CLONE_CHILD_SETTID\|SIGCHLD, child_tidptr=0x…) = 2000` que crea el proceso 2000; `t.2000` ejecuta `bash`, hace el `connect` a `AF_UNIX` de V53 y después la `execve` de `scripts/boe`: una sola invocación, la del proceso 2000 con el argv de `scripts/boe` y sin conexiones, porque el `connect` de `claude` es de otro proceso y el de `bash` es anterior a la `execve` del applet); `proceso-por-vfork` (`t.1000`, el proceso de `claude` en x86_64, con la `execve` de `claude` de `connect-fuera-de-la-invocacion`, la línea `vfork()` seguida de 33 espacios y `= 2000` y `+++ exited with 0 +++`; la segunda es la forma exacta con la que Claude Code 2.1.270 de x86_64 crea los procesos de sus órdenes en el runner, tal como la escribe `strace` 6.8 con su alineación por defecto (`-a 40`: el espacio tras el paréntesis de cierre y el relleno hasta la columna, con el `=` en la columna 41), registrada en la prueba de red del intento 3 de T030 (`gates/prueba-de-red.md`) y con el número del caso en lugar del real; `t.2000`, las cuatro líneas de `t.2000` de `connect-fuera-de-la-invocacion`: la `execve` de `bash -c`, el `connect` a `AF_UNIX` de V53, la `execve` de `scripts/boe` y `+++ exited with 0 +++`: una sola invocación, la del proceso 2000 con el argv `boe articulo BOE-A-2015-10565 a21 --json`, código 0 y sin conexiones, porque un proceso creado con `vfork` y con el relleno de alineación es el hilo principal de un proceso nuevo, data-model §9, regla 2); `connect-local-rechazado` (a `127.0.0.1:9` con `ECONNREFUSED`: `local`); `connect-publico-rechazado` (a `203.0.113.7:443` con `ENETUNREACH`, un error distinto de `EINPROGRESS`: `bloqueada`); `connect-publico-aceptado` (a `203.0.113.7:443` con resultado `0`: `red`); `connect-publico-en-curso` (a `203.0.113.7:443` con `EINPROGRESS`: `red`); `connect-ipv6-publico-en-curso` (`AF_INET6` a `[2001:db8::7]:443` con `EINPROGRESS`: `red`); `connect-af-unix` (`AF_UNIX`: `local`); `fichero-sin-origen` (`t.2000`, invocación de `scripts/boe`, y `t.2002`, con un `connect` a `203.0.113.7:443` con `EINPROGRESS`, sin que ninguna línea de los dos cree 2002: error que nombra `t.2000` y `t.2002` como ficheros sin la línea que los crea, en lugar de ignorar esa conexión); `fichero-ilegible` (una línea `execve` cortada, sin su resultado: error que nombra el fichero, el número de línea y su texto); `lineas-de-senal` (`t.1000`, un proceso de `bash` con su `execve`, que crea el proceso 2000 con la línea de `bash` de V53, recibe `--- SIGCHLD {si_signo=SIGCHLD, si_code=CLD_EXITED, si_pid=2000, si_uid=1001, si_status=5, si_utime=0, si_stime=1 /* 0.01 s */} ---` y termina en `+++ exited with 0 +++`; y `t.2000`, invocación de `scripts/boe` con `argv` `boe articulo BOE-A-2015-10565 a9998 --json`, con dos `--- SIGURG {si_signo=SIGURG, si_code=SI_TKILL, si_pid=2000, si_uid=1001} ---` antes de un `connect` a `127.0.0.1:9` con `EINPROGRESS` y `+++ exited with 5 +++`, las formas de V54 (2): sin error; las líneas de señal no cuentan, y queda una invocación con código 5 y una conexión `127.0.0.1:9` de clase `local`); `cortada-por-el-tope` (`t.2000`, invocación de `scripts/boe` con `argv` `boe articulo BOE-A-2015-10565 a9998 --json`, que crea el hilo 2001 con la línea de Go de `hilo-por-clone` y hace un `connect` a `203.0.113.7:443` con `EINPROGRESS`, **sin línea final**; y `t.2001`, **vacío**: como quedan los ficheros de los procesos vivos cuando `timeout` mata `strace` con `KILL`, research.md V54 (1) y (3): sin error; una invocación sin código (`Codigo` nil), con esa conexión atribuida y de clase `red`); `sin-linea-final-sin-corte` (los mismos `t.2000` y `t.2001`, con la sesión no cortada: traza ilegible, con un error que nombra `t.2000` y que le falta la línea final); `cortada-con-llamada-interrumpida` (`t.2000`, invocación de `scripts/boe` que crea el hilo 2001 con la línea de Go, recibe `--- SIGTERM {si_signo=SIGTERM, si_code=SI_USER, si_pid=1000, si_uid=1001} ---` y termina en `+++ killed by SIGTERM +++`; y `t.2001`, con `connect(9, {sa_family=AF_INET, sin_port=htons(443), sin_addr=inet_addr("203.0.113.7")}, 16) = ? ERESTARTSYS (To be restarted if SA_RESTART is set)`, esa misma línea de señal y `+++ killed by SIGTERM +++`, como deja un corte la llamada bloqueante que interrumpe, V54 (2): sin error; una invocación con código distinto de 0 y una conexión a `203.0.113.7:443` sin resultado, atribuida y de clase `red`); `llamada-interrumpida-sin-corte` (los mismos `t.2000` y `t.2001`, con la sesión no cortada: traza ilegible, con un error que nombra `t.2001`, la línea 1 y su texto); `llamada-interrumpida-antes-del-final` (`t.2000`, invocación de `scripts/boe` con un `connect` a `127.0.0.1:9` con `? ERESTARTSYS (To be restarted if SA_RESTART is set)` en la línea 2, seguido en la línea 3 de otro con `-1 ECONNREFUSED (Connection refused)` y de `+++ exited with 5 +++`, con la sesión cortada: traza ilegible, con un error que nombra `t.2000`, la línea 2 y su texto, porque a una llamada sin resultado solo pueden seguirla líneas de señal y la línea final) |
| `TestLeerTrazasSinFicheros` (`internal/evals/trazas_test.go`, sin etiqueta) | la regla 1 de data-model §9 sin ningún fichero y el error de un directorio que no se puede listar, sobre `t.TempDir()`, sin ningún fichero nuevo bajo `testdata/` (un directorio vacío no se versiona): `vacio` (un `traza/` vacío que crea el test: traza ilegible, con un error que nombra el directorio); `inexistente` (un directorio que no existe: error que lo nombra) |
| `TestInforme` | un subtest por directorio de `internal/evals/testdata/sesiones/informe/` (§9.1), con su nombre. A diferencia de `TestLeerSesion`, el directorio del caso no es una sesión sino una ejecución entera, un nivel por encima: lleva `evals/`, con las evals sintéticas como ficheros de datos, y `sesiones/<sesión>/`, un directorio por sesión como los que deja el guion (§3.2). El test llama a `EscribirInforme` con `Skill` `boe-legislacion`, `Evals` = `<caso>/evals`, `Sesiones` = `<caso>/sesiones`, `SinPython` = `informe/sin-python.txt`, `Destino` = un `t.TempDir()`, `Modelo` = `claude-haiku-4-5-20251001` y `Commit` = `0123456789abcdef0123456789abcdef01234567`, dos constantes del test, y lee `informe.json` e `informe.md` de `Destino`: `aprobado` (todas pasan: veredicto `aprobado`, `motivos` vacío, «ninguno» en ficheros mal formados, «ninguna» fuera de lo grabado, «ninguna petición llegó a la red de una fuente» y primera línea de `informe.md` igual a `# Informe de evals de boe-legislacion`; y la cabecera de data-model §10.3: en `informe.json`, `skill` `boe-legislacion`, `modelo` y `commit` iguales a las dos constantes, `modelos_de_sesion` igual a `["claude-haiku-4-5"]` —el de los transcripts de sus dos sesiones, una sola vez y distinto de la constante `Modelo` (§9.1), de modo que no pasan ni `modelo` tomado de las sesiones ni `modelos_de_sesion` tomado de la constante—, `versiones_de_claude_code` igual a `["2.1.270", "2.1.269"]`, las de `01-lpac-articulo-21` y `11-no-activa-programacion` en ese orden, y `sin_python` igual byte a byte al contenido de `informe/sin-python.txt`; y en `informe.md`, las líneas `Modelo del job: claude-haiku-4-5-20251001`, `Modelos de las sesiones: claude-haiku-4-5`, `Versiones de Claude Code: 2.1.270, 2.1.269` y `Commit: 0123456789abcdef0123456789abcdef01234567`, y la sección `## Comprobación sin Python` con el contenido de `informe/sin-python.txt` entero entre la línea ```` ```text ```` y la línea ```` ``` ```` (§5); y la `respuesta` de `01-lpac-articulo-21` en `informe.json` igual al `result` de su transcript sintético, que cita `[BOE-A-2015-10565, bloque a21]`, con `fin_de_la_sesion` `result success`, y la sección de esa sesión en `informe.md` con la pregunta de su `pregunta.txt` y esa respuesta (data-model §10.2)); `fuera-de-lo-grabado-no-cambia-el-veredicto` (las sesiones `01-lpac-articulo-21` y `01-lpac-articulo-21-prueba-de-red`, juzgadas las dos con `01-lpac-articulo-21.yaml`, pasan; las invocaciones de `a9998` de la segunda, con código 5 y 4, dan dos entradas de `fuera_de_lo_grabado` con `sesion` `01-lpac-articulo-21-prueba-de-red` y `eval` `01-lpac-articulo-21.yaml`, y dos filas con esa sesión y esa eval en la tabla sesión · eval · orden · código de `informe.md`; en `invocaciones` de `01-lpac-articulo-21-prueba-de-red`, la de `a9998 --json` con `conexiones` igual a `[{"destino": "127.0.0.1:9", "clase": "local"}]` —su traza tiene tres `connect` a esa dirección, como la real de research.md V53, y la pareja aparece una vez— y la de `--offline` con `conexiones` vacía; en la sección de esa sesión de `informe.md`, `127.0.0.1:9` y `local` en la invocación de `a9998 --json` y «sin conexiones» en la de `--offline`; veredicto `aprobado`); `fichero-mal-formado` (`LeerConjunto` da la eval 01 bien formada y `02-sin-pregunta.yaml` mal formada: una entrada de `ficheros_mal_formados` con `fichero` `02-sin-pregunta.yaml` y su error, y en `informe.md` la línea de ese fichero con su error; la sesión `01-lpac-articulo-21` se juzga con la 01 y pasa; `motivos` de la raíz exactamente una línea, `02-sin-pregunta.yaml: mal formado: <error>` con el mismo error de `ficheros_mal_formados`, y esa línea en los motivos de `informe.md`; veredicto `fallo`, por el fichero mal formado); `eval-que-no-pasa` (cita ausente: veredicto `fallo`, motivo precedido del nombre de la sesión); `sesion-sin-terminar` (la eval de no activación de §5, con código 124, sin `result`, con `activa` y `activada` iguales y con su `t.1000` sin línea final, que la regla 6 de data-model §9 admite en una sesión cortada: `pasa` falso, veredicto `fallo`, en la raíz el motivo precedido del nombre de la sesión, `11-no-activa-programacion: la sesión no terminó: tope de 240 s agotado (código 124)`, y ningún motivo `sesión ilegible`; `respuesta` vacía y `fin_de_la_sesion` `system`, el `type` de su último mensaje, el `init` (data-model §10.1); en `informe.md`, su motivo y su salida de error); `sesion-ilegible` (sin `codigo-de-la-sesion`: la eval no pasa con `sesión ilegible: codigo-de-la-sesion: …`, precedido del nombre de la sesión, y veredicto `fallo`; como su única sesión no se pudo leer, `modelos_de_sesion` y `versiones_de_claude_code` vacías, y en `informe.md` las líneas `Modelos de las sesiones: ninguno` y `Versiones de Claude Code: ninguna`); `llegada-a-la-red` (conexión de clase `red` de la invocación `a9998 --json` de la sesión `01-lpac-articulo-21`: una entrada de `red` con `sesion` `01-lpac-articulo-21`, `eval` `01-lpac-articulo-21.yaml`, la orden y el destino `203.0.113.7:443`, su fila en la tabla sesión · eval · orden · destino de `informe.md`; `motivos` de la raíz exactamente `["01-lpac-articulo-21: petición llegada a la red: boe articulo BOE-A-2015-10565 a9998 --json → 203.0.113.7:443"]`, y esa línea en los motivos de `informe.md`; y veredicto `fallo` aunque la eval pase); `sesion-cortada-con-invocaciones` (la sesión `01-lpac-articulo-21` con código 137, sin `result`, con `t.1000` sin línea final, la invocación `a9998 --offline --json` de `t.2000`, que terminó antes del corte con código 4, y la invocación `a9998 --json` de `t.2001`, sin línea final y con un `connect` a `203.0.113.7:443` con `EINPROGRESS`: la eval no pasa, con los motivos `la sesión no terminó: terminada por señal tras el tope (código 137)`, el del comando ausente y el de la cita ausente, y ninguno `sesión ilegible`; en `invocaciones`, primero la de `--offline` con `codigo` 4 y `conexiones` vacía y después la de `--json` con `codigo` `null` y `conexiones` igual a `[{"destino": "203.0.113.7:443", "clase": "red"}]`; una entrada de `fuera_de_lo_grabado` con la orden `boe articulo BOE-A-2015-10565 a9998 --offline --json` y código 4, y ninguna con la otra; una de `red` con la orden `boe articulo BOE-A-2015-10565 a9998 --json` y el destino `203.0.113.7:443`; `otras_fallidas` vacía; en `informe.md`, «sin código (sesión cortada)» en esa invocación y sus filas en las tablas de fuera de lo grabado y de red; veredicto `fallo`); `eval-sin-sesion` (`evals/` con `01-lpac-articulo-21.yaml` y `11-no-activa-programacion.yaml`, y solo la sesión `01-lpac-articulo-21`, igual que en `aprobado`: un único resultado en `evals`, el de esa sesión, con `pasa` verdadero; `motivos` de la raíz exactamente `["11-no-activa-programacion.yaml: sin ninguna sesión"]`, y esa línea en los motivos de `informe.md`; veredicto `fallo`, que sin la regla del paso 3 de §3.3 saldría `aprobado` sin haber evaluado la 11); `sin-eval-txt` (la sesión `01-lpac-articulo-21` de `aprobado` sin `eval.txt`: su resultado con `eval` vacío, `pasa` falso y el motivo `sesión ilegible: eval.txt: …`; `motivos` de la raíz, en este orden, `01-lpac-articulo-21: sesión ilegible: eval.txt: …` y `01-lpac-articulo-21.yaml: sin ninguna sesión`, porque ningún `eval.txt` la nombra; veredicto `fallo`); `eval-desconocida` (la misma sesión con `eval.txt` igual a `03-inexistente.yaml`, que no es ninguna eval de `evals/`: su resultado con `eval` `03-inexistente.yaml`, `pasa` falso y el motivo `sesión ilegible: eval.txt: …`, que nombra `03-inexistente.yaml`; en la raíz, ese motivo precedido del nombre de la sesión y después `01-lpac-articulo-21.yaml: sin ninguna sesión`; veredicto `fallo`); `sin-pregunta-txt` (la misma sesión sin `pregunta.txt` y con su `eval.txt`: `pasa` falso con el único motivo `sesión ilegible: pregunta.txt: …`, precedido en la raíz del nombre de la sesión, y ningún motivo `sin ninguna sesión`, porque su `eval.txt` nombra la 01; veredicto `fallo`); `traza-ilegible` (la sesión `01-lpac-articulo-21` de `aprobado` con su `traza/t.2000` cuya línea `execve` está cortada sin su resultado, la forma de `TestLeerTrazas/fichero-ilegible`: `pasa` falso con un único motivo que empieza por `sesión ilegible: traza` y nombra `t.2000`, la línea y su texto, precedido en la raíz del nombre de la sesión; ningún motivo `sin ninguna sesión`, porque su `eval.txt` nombra la 01; veredicto `fallo`) |
| `TestEscribirInformeSinSusEntradas` (`internal/evals/informe_test.go`, sin etiqueta) | el contrato de error de `EscribirInforme` (§3.3) y el informe sin nada que juzgar, en una tabla sobre directorios temporales, sin ningún fichero nuevo bajo `testdata/`. Cada caso parte de las entradas del caso `aprobado` de `TestInforme` (§9.1), que solo lee —`Evals` = sus `evals/`, `Sesiones` = sus `sesiones/`, `SinPython` = `informe/sin-python.txt`—, con `Destino` = un `t.TempDir()` vacío y las dos constantes de `TestInforme`, y cambia una sola: `sin-python-inexistente` (`SinPython` = una ruta inexistente dentro de otro `t.TempDir()`: error que nombra esa ruta, y en `Destino` ni `informe.md` ni `informe.json`); `sesiones-inexistente` (`Sesiones` = un directorio inexistente: error que lo nombra, y nada en `Destino`); `evals-inexistente` (`Evals` = un directorio inexistente: error que lo nombra, y nada en `Destino`); `destino-es-un-fichero` (`Destino` = un fichero regular que el test crea con `0o600` y un contenido constante: error que nombra esa ruta, y el fichero con el mismo contenido); `evals-y-sesiones-vacios` (`Evals` y `Sesiones` = dos directorios vacíos que el test crea: sin error; `informe.json` e `informe.md` escritos en `Destino`, con `evals` vacía, `motivos` exactamente `["ninguna eval bien formada que juzgar"]`, esa línea en los motivos de `informe.md` y veredicto `fallo`). Los dos directorios vacíos de este caso, como el `Destino` y las rutas inexistentes de los demás, los crea el propio test en un `t.TempDir()`: ningún caso de esta tabla añade ficheros al árbol de §9.1 |
| `TestPrepararDirectorioDeSesion` (`internal/evals/preparar_test.go`, sin etiqueta) | `PrepararSesion` (§3.2) sobre las grabaciones de H4 y evals sintéticas que el propio test escribe desde constantes en un `t.TempDir()` por subtest —el paso del plan que lo introduce es de código y no hay evals sintéticas en `testdata/` hasta las sesiones—: `01-lpac-articulo-21.yaml` (pregunta «¿qué dice el art. 21 de la Ley 39/2015?», bloque y cita `BOE-A-2015-10565` `a21`) y `02-lpac-articulo-22.yaml` (pregunta «¿qué dice el art. 22 de la Ley 39/2015?», bloque y cita `BOE-A-2015-10565` `a22`), las dos con sus respuestas en las grabaciones de H4, en todos los subtests; y, solo donde su subtest lo dice, `03-sin-pregunta.yaml` (la 01 sin la clave `pregunta`) y `03-lpac-articulo-9998.yaml` (pregunta «¿qué dice el art. 9998 de la Ley 39/2015?», bloque y cita `BOE-A-2015-10565` `a9998`, que no está en ninguna grabación). El directorio de la sesión es otro `t.TempDir()` con `cache/` creado vacío. Subtests: `eval-normal` (`Fichero` `02-lpac-articulo-22.yaml`: ni faltas ni error; `pregunta.txt` es la pregunta de la 02 y un salto de línea; `eval.txt`, `02-lpac-articulo-22.yaml` y un salto de línea; y `ComprobarSinRed` sobre ese `cache/` con `ConsultasNecesarias` de las **dos** evals no da faltas, `boe articulo BOE-A-2015-10565 a21` de la otra eval incluida); `prueba-de-red` (`Fichero` `01-lpac-articulo-21.yaml` y `PruebaDeRed`: `pregunta.txt` es exactamente la pregunta de la 01, una línea en blanco, el texto literal de §6 con las dos órdenes `a9998` y un salto de línea; `eval.txt`, `01-lpac-articulo-21.yaml` y un salto de línea); `eval-inexistente` (`Fichero` `03-inexistente.yaml`: error que nombra `03-inexistente.yaml`, ninguna falta, y ni `pregunta.txt` ni `eval.txt` en el directorio); `eval-mal-formada` (01, 02 y `03-sin-pregunta.yaml`, con `Fichero` `02-lpac-articulo-22.yaml`: error que nombra `03-sin-pregunta.yaml` con su error, ninguna falta, `cache/` sigue vacío y ni `pregunta.txt` ni `eval.txt` en el directorio: paso 1 de §3.2, sin preparar ni escribir nada); `con-faltas` (01, 02 y `03-lpac-articulo-9998.yaml`, con `Fichero` `01-lpac-articulo-21.yaml`: sin error, faltas que nombran `03-lpac-articulo-9998.yaml` y `boe articulo BOE-A-2015-10565 a9998`, y ni `pregunta.txt` ni `eval.txt` en el directorio: paso 3 de §3.2) |
| `TestPrepararSesion`, `TestInformeDelJob` | etiqueta `evals`; solo leen banderas y llaman a `PrepararSesion` y `EscribirInforme`, que fijan `TestPrepararDirectorioDeSesion` y `TestInforme` |

### 9.1 Árbol de `internal/evals/testdata/sesiones/`

Un subdirectorio por test, porque cada función lee un nivel distinto: `LeerSesion`, el directorio de una sesión;
`LeerTrazas`, el `traza/` de una sesión; `EscribirInforme`, una ejecución entera con sus evals y sus sesiones. Son
ficheros nuevos bajo `testdata/` fuera de `internal/source/`, sin pausa (research.md V36), en tres tareas `[datos]`, una
por subdirectorio, y el caso `proceso-por-vfork` de `leer-trazas/` en una cuarta, posterior a la prueba de red, que
nombran cada fichero por su ruta completa: `internal/evals/testdata/sesiones/` seguido de la ruta de este árbol (p. ej.
`internal/evals/testdata/sesiones/leer-sesion/activada/sesion.jsonl`). No hay ningún otro fichero: 50 casos (15, 22 y
13) y 197 ficheros (42, 34 y 121). Cada fichero de traza tiene solo líneas con las formas de research.md V53, las de
señal de V54 donde su caso las nombra y, en `proceso-por-vfork`, la línea `vfork()` con el relleno de alineación que
registró la prueba de red (§9), y termina en su línea final (data-model §9), salvo en
los casos de corte, que tienen lo que puede dejar un corte (research.md V54; data-model §9, regla 6) tal como los
describe §9: `cortada-por-el-tope`, `sin-linea-final-sin-corte`, `cortada-con-llamada-interrumpida`,
`llamada-interrumpida-sin-corte` y `llamada-interrumpida-antes-del-final` en `leer-trazas/`, y `sesion-sin-terminar` y
`sesion-cortada-con-invocaciones` en `informe/`. En `informe/`, `t.1000` es el proceso de `claude`: su
`execve`, una línea `clone(child_stack=NULL, flags=CLONE_CHILD_CLEARTID|CLONE_CHILD_SETTID|SIGCHLD, child_tidptr=0x…) = <n>`
por cada otro fichero de la traza de su sesión, que es la que lo crea, y su línea final, salvo en esos dos casos de
corte, en los que no la tiene; `t.2000`, `t.2001` y `t.2002` son
invocaciones de `/home/runner/.claude/skills/boe-legislacion/scripts/boe`. En `leer-trazas/`, qué es cada fichero lo dice
su caso en §9.

```text
leer-sesion/                                      TestLeerSesion: el directorio del caso es el de una sesión
  activada/                                       sesion.jsonl  codigo-de-la-sesion  sesion.err
  no-activada/                                    sesion.jsonl  codigo-de-la-sesion  sesion.err
  otra-skill-activada/                            sesion.jsonl  codigo-de-la-sesion  sesion.err
  codigo-distinto-de-cero/                        sesion.jsonl  codigo-de-la-sesion  sesion.err
  tope-agotado/                                   sesion.jsonl  codigo-de-la-sesion  sesion.err
  senal-tras-el-tope/                             sesion.jsonl  codigo-de-la-sesion  sesion.err
  sin-result/                                     sesion.jsonl  codigo-de-la-sesion  sesion.err
  error-max-turns/                                sesion.jsonl  codigo-de-la-sesion  sesion.err
  result-con-is-error/                            sesion.jsonl  codigo-de-la-sesion  sesion.err
  sin-fichero-de-codigo/                          sesion.jsonl  sesion.err
  codigo-no-entero/                               sesion.jsonl  codigo-de-la-sesion  sesion.err
  sin-transcript/                                 codigo-de-la-sesion  sesion.err
  sin-salida-de-error/                            sesion.jsonl  codigo-de-la-sesion
  linea-ilegible/                                 sesion.jsonl  codigo-de-la-sesion  sesion.err
  sin-mensajes/                                   sesion.jsonl  codigo-de-la-sesion  sesion.err
leer-trazas/                                      TestLeerTrazas: el directorio del caso es el traza/ de una sesión
  argv-escapado/                                  t.2000
  salida-con-codigo/                              t.2000
  muerte-por-senal/                               t.2000
  hilo-por-clone/                                 t.2000  t.2001
  hilo-por-clone3/                                t.2000  t.2001
  hilo-de-un-hilo/                                t.2000  t.2001  t.2002
  connect-fuera-de-la-invocacion/                 t.1000  t.2000
  proceso-por-vfork/                              t.1000  t.2000
  connect-local-rechazado/                        t.2000
  connect-publico-rechazado/                      t.2000
  connect-publico-aceptado/                       t.2000
  connect-publico-en-curso/                       t.2000
  connect-ipv6-publico-en-curso/                  t.2000
  connect-af-unix/                                t.2000
  fichero-sin-origen/                             t.2000  t.2002
  fichero-ilegible/                               t.2000
  lineas-de-senal/                                t.1000  t.2000
  cortada-por-el-tope/                            t.2000  t.2001
  sin-linea-final-sin-corte/                      t.2000  t.2001
  cortada-con-llamada-interrumpida/               t.2000  t.2001
  llamada-interrumpida-sin-corte/                 t.2000  t.2001
  llamada-interrumpida-antes-del-final/           t.2000
informe/                                          TestInforme: el directorio del caso es una ejecución entera
  sin-python.txt                                  el mismo para los trece casos
  aprobado/
    evals/                                        01-lpac-articulo-21.yaml  11-no-activa-programacion.yaml
    sesiones/01-lpac-articulo-21/                 eval.txt  pregunta.txt  sesion.jsonl  codigo-de-la-sesion  sesion.err  traza/t.1000  traza/t.2000
    sesiones/11-no-activa-programacion/           eval.txt  pregunta.txt  sesion.jsonl  codigo-de-la-sesion  sesion.err  traza/t.1000
  fuera-de-lo-grabado-no-cambia-el-veredicto/
    evals/                                        01-lpac-articulo-21.yaml
    sesiones/01-lpac-articulo-21/                 eval.txt  pregunta.txt  sesion.jsonl  codigo-de-la-sesion  sesion.err  traza/t.1000  traza/t.2000
    sesiones/01-lpac-articulo-21-prueba-de-red/   eval.txt  pregunta.txt  sesion.jsonl  codigo-de-la-sesion  sesion.err  traza/t.1000  traza/t.2000  traza/t.2001  traza/t.2002
  fichero-mal-formado/
    evals/                                        01-lpac-articulo-21.yaml  02-sin-pregunta.yaml
    sesiones/01-lpac-articulo-21/                 eval.txt  pregunta.txt  sesion.jsonl  codigo-de-la-sesion  sesion.err  traza/t.1000  traza/t.2000
  eval-que-no-pasa/
    evals/                                        01-lpac-articulo-21.yaml
    sesiones/01-lpac-articulo-21/                 eval.txt  pregunta.txt  sesion.jsonl  codigo-de-la-sesion  sesion.err  traza/t.1000  traza/t.2000
  sesion-sin-terminar/
    evals/                                        11-no-activa-programacion.yaml
    sesiones/11-no-activa-programacion/           eval.txt  pregunta.txt  sesion.jsonl  codigo-de-la-sesion  sesion.err  traza/t.1000
  sesion-ilegible/
    evals/                                        01-lpac-articulo-21.yaml
    sesiones/01-lpac-articulo-21/                 eval.txt  pregunta.txt  sesion.jsonl  sesion.err  traza/t.1000  traza/t.2000
  llegada-a-la-red/
    evals/                                        01-lpac-articulo-21.yaml
    sesiones/01-lpac-articulo-21/                 eval.txt  pregunta.txt  sesion.jsonl  codigo-de-la-sesion  sesion.err  traza/t.1000  traza/t.2000  traza/t.2001
  sesion-cortada-con-invocaciones/
    evals/                                        01-lpac-articulo-21.yaml
    sesiones/01-lpac-articulo-21/                 eval.txt  pregunta.txt  sesion.jsonl  codigo-de-la-sesion  sesion.err  traza/t.1000  traza/t.2000  traza/t.2001
  eval-sin-sesion/
    evals/                                        01-lpac-articulo-21.yaml  11-no-activa-programacion.yaml
    sesiones/01-lpac-articulo-21/                 eval.txt  pregunta.txt  sesion.jsonl  codigo-de-la-sesion  sesion.err  traza/t.1000  traza/t.2000
  sin-eval-txt/
    evals/                                        01-lpac-articulo-21.yaml
    sesiones/01-lpac-articulo-21/                 pregunta.txt  sesion.jsonl  codigo-de-la-sesion  sesion.err  traza/t.1000  traza/t.2000
  eval-desconocida/
    evals/                                        01-lpac-articulo-21.yaml
    sesiones/01-lpac-articulo-21/                 eval.txt  pregunta.txt  sesion.jsonl  codigo-de-la-sesion  sesion.err  traza/t.1000  traza/t.2000
  sin-pregunta-txt/
    evals/                                        01-lpac-articulo-21.yaml
    sesiones/01-lpac-articulo-21/                 eval.txt  sesion.jsonl  codigo-de-la-sesion  sesion.err  traza/t.1000  traza/t.2000
  traza-ilegible/
    evals/                                        01-lpac-articulo-21.yaml
    sesiones/01-lpac-articulo-21/                 eval.txt  pregunta.txt  sesion.jsonl  codigo-de-la-sesion  sesion.err  traza/t.1000  traza/t.2000
```

Contenido de lo que hay bajo `informe/` (el de `leer-sesion/` y `leer-trazas/` es el de cada subtest en la tabla de
arriba):

| Fichero o sesión | Casos | Contenido |
|---|---|---|
| `evals/01-lpac-articulo-21.yaml` | todos salvo `sesion-sin-terminar` | la eval positiva de ejemplo del contrato de evals §1 |
| `evals/11-no-activa-programacion.yaml` | `aprobado`, `sesion-sin-terminar`, `eval-sin-sesion` | la eval de no activación de ejemplo del contrato de evals §1 |
| `evals/02-sin-pregunta.yaml` | `fichero-mal-formado` | la eval 01 sin la clave `pregunta` |
| `sin-python.txt` | todos | exactamente las tres líneas que escribe la comprobación 3 de §3.1 sin ningún Python encontrado, tal como las da en research.md V56: `búsqueda: ` seguido de la orden de búsqueda como root en todo el sistema de ficheros salvo `/proc` y `/sys`, con sus argumentos separados por un espacio (el texto literal de §3.1), `usuario: root` y `resultado: ninguno`, cada una con su salto de línea |
| mensaje `init` del transcript de cada sesión | todos | `model` `claude-haiku-4-5` y `claude_code_version` `2.1.270`, salvo en `11-no-activa-programacion` de `aprobado`, que lleva `claude_code_version` `2.1.269`. El modelo es a propósito distinto de la constante `Modelo` (`claude-haiku-4-5-20251001`) que `TestInforme` pasa a `EscribirInforme` (§9), para que no pasen ni `modelo` tomado de las sesiones ni `modelos_de_sesion` tomado de la constante; las dos sesiones de `aprobado` con el mismo modelo fijan que se da una vez, y sus dos versiones, el orden por nombre de sesión (data-model §10.3) |
| `eval.txt` y `pregunta.txt` de cada sesión | todos, salvo lo que dicen de ellos las filas de `sin-eval-txt`, `eval-desconocida` y `sin-pregunta-txt` | el nombre de la eval con la que se juzga y su pregunta, cada uno con un salto de línea final; en `01-lpac-articulo-21-prueba-de-red`, `01-lpac-articulo-21.yaml` y la pregunta con el texto de §6, como los escribe `PrepararSesion` |
| sesión `01-lpac-articulo-21` | `aprobado`, `fuera-de-lo-grabado-no-cambia-el-veredicto`, `fichero-mal-formado`, `llegada-a-la-red`, `eval-sin-sesion` | código `0`; transcript con `init`, `tool_use` `Skill` de `boe-legislacion` y `result` `success` cuya respuesta cita `[BOE-A-2015-10565, bloque a21]`; `sesion.err` vacío; `t.2000`: `boe articulo BOE-A-2015-10565 a21 --json` con código 0; en `llegada-a-la-red`, además `t.2001`: `boe articulo BOE-A-2015-10565 a9998 --json` con un `connect` a `203.0.113.7:443` con `EINPROGRESS` y código 5 |
| sesión `01-lpac-articulo-21` | `eval-que-no-pasa` | como en `aprobado`, pero la respuesta cita `[BOE-A-2015-10565, bloque a22]` y no `a21` |
| sesión `01-lpac-articulo-21` | `sesion-ilegible` | como en `aprobado`, sin `codigo-de-la-sesion` |
| sesión `01-lpac-articulo-21` | `sin-eval-txt` | como en `aprobado`, sin `eval.txt` |
| sesión `01-lpac-articulo-21` | `eval-desconocida` | como en `aprobado`, con `eval.txt` igual a `03-inexistente.yaml` y un salto de línea |
| sesión `01-lpac-articulo-21` | `sin-pregunta-txt` | como en `aprobado`, sin `pregunta.txt` |
| sesión `01-lpac-articulo-21` | `traza-ilegible` | como en `aprobado`, con la línea `execve` de `t.2000` cortada sin su resultado, como en `TestLeerTrazas/fichero-ilegible` |
| sesión `01-lpac-articulo-21-prueba-de-red` | `fuera-de-lo-grabado-no-cambia-el-veredicto` | como `01-lpac-articulo-21` de `aprobado`, y además `t.2001`: `boe articulo BOE-A-2015-10565 a9998 --json` con tres `connect` a `127.0.0.1:9` con resultado `-1 EINPROGRESS (Operation now in progress)`, como en la traza real de research.md V53, y código 5, y `t.2002`: `boe articulo BOE-A-2015-10565 a9998 --offline --json` sin ningún `connect` y con código 4 |
| sesión `11-no-activa-programacion` | `aprobado` | código `0`; transcript con `init` y `result` `success`, sin `tool_use` `Skill`; `sesion.err` vacío; solo `t.1000` |
| sesión `11-no-activa-programacion` | `sesion-sin-terminar` | código `124`; transcript con solo el mensaje `init`; `sesion.err` con una línea de texto; solo `t.1000`, sin línea final (data-model §9, regla 6) |
| sesión `01-lpac-articulo-21` | `sesion-cortada-con-invocaciones` | código `137`; transcript con `init` y `tool_use` `Skill` de `boe-legislacion`, sin `result`; `sesion.err` con una línea de texto; `t.1000` sin línea final; `t.2000`: `boe articulo BOE-A-2015-10565 a9998 --offline --json` sin ningún `connect` y con `+++ exited with 4 +++`; `t.2001`: `boe articulo BOE-A-2015-10565 a9998 --json` con un `connect` a `203.0.113.7:443` con `EINPROGRESS` y sin línea final |
