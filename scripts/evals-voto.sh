#!/usr/bin/env bash
# Abre un voto del juez con modelo de las evals con Claude Code (contracts/juez-y-voto.md §4 de H24; research.md D4 de
# H24; ADR 0037). Lo ejecuta el votante de internal/evals, sin argumentos, en el directorio cwd/ del directorio del juez
# que acaba de crear, con el mensaje del voto por la entrada estándar: lee el modelo, la rúbrica y el esquema de
# ../modelo.txt, ../rubrica.md y ../esquema.json y ejecuta con exec la orden de la validación del juez, la de voto_real
# de evidencias/adr-0037/guiones/juez.py, argumento a argumento y en su orden: una sesión sin herramientas, sin
# servidores MCP, sin órdenes de barra y sin guardar, con la rúbrica como instrucciones y el esquema como forma de la
# respuesta. El tope de 35 s, el entorno —el PATH, cuyo primer claude es el del juez, HOME, CLAUDE_CONFIG_DIR y la
# credencial, y nada más— y la lectura de la salida los pone el votante. No lee ninguna variable. No usa nada posterior
# a bash 3.2, el de macOS.
#
# Abre una sesión con modelo y consume la credencial de Claude Code: solo lo ejecuta el votante, en el job de evals, en
# la ejecución de la medida del juez y en el sondeo; ni make ci, ni los ganchos, ni ninguna tarea del workflow. Los
# tests de internal/evals lo ejecutan con un claude sustituto delante en el PATH, que no abre ninguna sesión.
set -euo pipefail

modelo="$(cat ../modelo.txt)"

# La rúbrica va entera, con su salto de línea final, como en la validación; $(…) quita los saltos de línea finales, así
# que se lee con un carácter detrás, que luego se quita (research.md V22 de H24). Con &&, la lectura que falla detiene
# el guion: dentro de $(…) bash no hereda -e.
rubrica="$(cat ../rubrica.md && printf x)"
rubrica="${rubrica%x}"

# El esquema va sin su salto de línea final, también como en la validación.
esquema="$(cat ../esquema.json)"

exec claude -p --model "$modelo" --tools "" --strict-mcp-config --disable-slash-commands --no-session-persistence \
	--system-prompt "$rubrica" --output-format json --json-schema "$esquema"
