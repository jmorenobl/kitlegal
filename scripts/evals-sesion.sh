#!/usr/bin/env bash
# Abre una sesión de evals con Claude Code (contracts/ejecucion-del-job.md §2 de H7.3; research.md D8 de H7.3). Lo
# ejecuta el repartidor de internal/evals, sin argumentos, en el directorio trabajo/ de la sesión que acaba de preparar:
# lee la pregunta y el modelo de ../pregunta.txt y ../modelo.txt y ejecuta con exec la orden de la sesión de siempre,
# carácter a carácter. Con KITLEGAL_EVALS_TRAZA=si, el job, la ejecuta bajo strace, con la traza en ../traza/. El grupo
# de procesos, el tope de 240 s, el entorno —la caché, el proxy que rechaza toda petición salvo la del modelo, el estado
# de Claude Code y el temporal de la sesión— y los ficheros de su salida los pone el repartidor. No usa nada posterior
# a bash 3.2, el de macOS.
#
# Abre una sesión con modelo y consume la credencial de Claude Code: solo lo ejecuta el repartidor, en el job de evals y
# en el sondeo; ni make ci, ni los ganchos, ni ninguna tarea del workflow.
set -euo pipefail

pregunta="$(cat ../pregunta.txt)"
modelo="$(cat ../modelo.txt)"

# Sin el aislamiento de subprocesos de Claude Code, que en Linux ejecuta cada orden dentro de bwrap, con un espacio de
# nombres de PID que deja la traza sin atribuir y el disco de solo lectura; Claude Code sigue sin pasar la credencial
# del modelo al entorno de las órdenes (research.md V12, V61 y D13 de H5).
sesion=(claude -p "$pregunta" --model "$modelo" --output-format stream-json --verbose --max-turns 30
	--no-session-persistence --setting-sources user --settings '{"sandbox":{"enabled":false}}'
	--permission-mode bypassPermissions --disallowedTools WebFetch WebSearch)

# La traza lleva entera cada cadena de un execve: -s 131072 es el tamaño máximo de un argumento en Linux, y con -s 4096
# la instantánea de shell que Claude Code crea antes de la primera orden de Bash sale cortada y deja ilegible la traza
# (research.md V61 y V62 de H5).
if [[ "${KITLEGAL_EVALS_TRAZA:-}" == si ]]; then
	exec strace -ff -e trace=execve,connect,clone,clone3,fork,vfork -s 131072 -o ../traza/t -- "${sesion[@]}"
fi

exec "${sesion[@]}"
