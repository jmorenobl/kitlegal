#!/bin/sh
trap 'printf "install.sh: no se ha instalado nada: el guion no llegó entero (¿se cortó la descarga?)\n" >&2; exit 1' EXIT
# Instala el binario kitlegal sin clonar el repositorio: descarga de las releases
# de GitHub el archivo de esta plataforma y checksums.txt, comprueba la huella
# SHA-256 del archivo y deja kitlegal en $KITLEGAL_INSTALL_DIR o, si no está
# definido o está vacío, en $HOME/.local/bin (FR-100 a FR-107;
# specs/009-h19-instalar-sin-clonar/contracts/release.md §7; research D22, D27).
#
#   curl -fsSL https://raw.githubusercontent.com/jmorenobl/kitlegal/main/scripts/install.sh | sh
#   curl -fsSL …/install.sh | sh -s -- <versión>
#   sh install.sh [<versión>]
#
# Sin versión instala la última release; con ella, la de la etiqueta
# v<versión>: 0.1.0 y v0.1.0 piden la misma. Es POSIX sh y no necesita bash, Go,
# git ni ningún gestor de paquetes: además de las utilidades POSIX básicas
# (uname, grep, awk, mkdir, cp, chmod, mv, rm), solo curl, tar, mktemp y
# sha256sum o shasum. No toca ningún fichero de arranque del shell ni nada
# fuera del directorio de instalación: si ese directorio no está en el PATH,
# imprime la línea que lo añade. Un kitlegal anterior solo se sustituye al
# final, de una vez, y queda intacto ante cualquier error.
#
# KITLEGAL_INSTALL_URL cambia la base de las descargas; las pruebas la apuntan a
# un origen local file://…, sin red (FR-107).
#
# Un guion que llega cortado por `curl … | sh` no ejecuta nada a medias
# (FR-100, FR-106): todo lo que hace está en funciones, y lo único que se
# ejecuta por sí mismo es la guarda de la segunda línea y, en la última, la
# llamada a main, dentro de un grupo que sh no ejecuta si le falta la llave que
# lo cierra. Si el corte deja las funciones enteras pero no la llamada, la
# guarda, que main retira al empezar, termina con una línea con el prefijo
# install.sh: y código 1, en lugar de salir con 0 sin haber instalado nada.
# Solo un corte antes de que llegue la guarda entera sale con 0, igual que un
# guion vacío: sh no ha recibido nada que ejecutar.
#
# Cada variable seguida de un carácter que no es ASCII (las comillas «») va
# entre llaves: el /bin/sh de macOS, bash 3.2, en un locale UTF-8 lee los bytes
# de ese carácter como parte del nombre y, con set -u, aborta por una variable
# sin definir en lugar de dar el mensaje de fallo.
set -eu

# fallo termina con una sola línea en la salida de error, con el prefijo
# install.sh:, la versión pedida y lo que falló, y código 1. Lo llama todo error:
# ninguno llega después de sustituir kitlegal, así que no se ha instalado nada.
fallo() {
	printf 'install.sh: no se pudo instalar %s: %s\n' "$que" "$1" >&2
	exit 1
}

# citar deja en citado el texto entre comillas simples de sh, con cada comilla
# simple escrita '\'', sin salir del shell: vale para cualquier nombre de
# directorio.
citar() {
	resto=$1
	citado=
	while :; do
		case $resto in
		*\'*)
			citado=$citado${resto%%\'*}\'\\\'\'
			resto=${resto#*\'}
			;;
		*)
			citado=\'$citado$resto\'
			return
			;;
		esac
	done
}

# huella escribe la huella SHA-256 del fichero, con sha256sum o, si no está,
# con shasum -a 256. Lo lee por la entrada estándar para que su nombre no
# cambie la forma de la salida.
huella() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum <"$1"
	else
		shasum -a 256 <"$1"
	fi
}

# limpiar retira los temporales: tmp, el de mktemp -d, que recibe las descargas
# y lo extraído, y provisional, el del directorio de instalación, la copia que
# se renombra encima de kitlegal y que, después de renombrarla, ya no existe
# (FR-104).
limpiar() {
	if [ -n "$tmp" ]; then
		rm -rf "$tmp"
	fi
	if [ -n "$provisional" ]; then
		rm -f "$provisional"
	fi
}

# main instala la versión que piden sus argumentos, ninguno o uno.
main() {
	# El guion llegó entero: la guarda de la segunda línea ya no hace falta.
	trap - EXIT

	base=${KITLEGAL_INSTALL_URL:-https://github.com/jmorenobl/kitlegal/releases}

	# semver es la gramática de SemVer 2.0.0 en ERE, con una v minúscula inicial
	# opcional: la misma que FormaSemVer de internal/core/instalacion (research
	# D31), con [0-9] en lugar de \d; se evalúa con LC_ALL=C.
	semver='^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-(0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*)(\.(0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*))*)?(\+[0-9a-zA-Z-]+(\.[0-9a-zA-Z-]+)*)?$'

	nl='
'

	# que es la versión pedida tal como la nombra cada error: la última sin
	# argumento, o la del argumento.
	que='la última versión'

	# --- La versión: 0 o 1 argumentos (FR-100) ---

	if [ "$#" -gt 1 ]; then
		que=
		for argumento in "$@"; do
			que="${que} «${argumento}»"
		done
		que=${que# }
		fallo "se admite como mucho un argumento, la versión, y se recibieron $#"
	fi

	if [ "$#" -eq 1 ]; then
		# grep mira línea a línea: una versión con un salto de línea se rechaza
		# antes, y sin copiarla, para que el error siga siendo una sola línea.
		case $1 in
		*"$nl"*)
			que='la versión pedida'
			fallo "tiene un salto de línea: no tiene la forma de SemVer 2.0.0 (por ejemplo, 0.1.0 o v0.1.0)"
			;;
		esac
		que="la versión «$1»"
		if ! printf '%s\n' "$1" | LC_ALL=C grep -Eq "$semver"; then
			fallo "no tiene la forma de SemVer 2.0.0 (por ejemplo, 0.1.0 o v0.1.0)"
		fi
		version=${1#v}
		que="la versión v$version"
		descarga=download/v$version
	else
		descarga=latest/download
	fi

	# --- El sistema y la arquitectura (FR-101) ---

	sistema=$(uname -s) || fallo "no se pudo leer el sistema con uname -s"
	case $sistema in
	Darwin) so=darwin ;;
	Linux) so=linux ;;
	*) fallo "el sistema «${sistema}» (uname -s) no se admite: solo Darwin y Linux" ;;
	esac

	maquina=$(uname -m) || fallo "no se pudo leer la arquitectura con uname -m"
	case $maquina in
	x86_64 | amd64) arq=amd64 ;;
	arm64 | aarch64) arq=arm64 ;;
	*) fallo "la arquitectura «${maquina}» (uname -m) no se admite: solo x86_64, amd64, arm64 y aarch64" ;;
	esac

	# --- El directorio de instalación, antes de descargar nada (FR-103) ---
	#
	# Sin KITLEGAL_INSTALL_DIR, $HOME/.local/bin, y nunca /.local/bin ni una ruta
	# relativa: HOME tiene que ser una ruta absoluta que no sea la raíz.

	if [ -n "${KITLEGAL_INSTALL_DIR:-}" ]; then
		dir=$KITLEGAL_INSTALL_DIR
	else
		sinDirectorio="no hay directorio de instalación: define KITLEGAL_INSTALL_DIR"
		case ${HOME:-} in
		'') fallo "HOME no está definido o está vacío, y KITLEGAL_INSTALL_DIR tampoco; $sinDirectorio" ;;
		/*[!/]*) ;;
		/*) fallo "HOME es la raíz del sistema y KITLEGAL_INSTALL_DIR no está definido o está vacío; $sinDirectorio" ;;
		*) fallo "HOME («${HOME}») no es una ruta absoluta y KITLEGAL_INSTALL_DIR no está definido o está vacío; $sinDirectorio" ;;
		esac
		dir=$HOME/.local/bin
	fi

	# ruta es dir tal como se pasa a las órdenes: una relativa con ./ delante,
	# para que ninguna empiece por - y se lea como opción.
	case $dir in
	/*) ruta=$dir ;;
	*) ruta=./$dir ;;
	esac

	if ! command -v curl >/dev/null 2>&1; then
		fallo "no se encuentra curl, que hace falta para descargar"
	fi
	if ! command -v sha256sum >/dev/null 2>&1 && ! command -v shasum >/dev/null 2>&1; then
		fallo "no se encuentra sha256sum ni shasum, que hacen falta para comprobar la huella"
	fi

	# --- Los temporales, retirados siempre (FR-104) ---

	tmp=
	provisional=
	trap limpiar EXIT
	trap 'exit 1' HUP INT TERM

	tmp=$(mktemp -d) || fallo "no se pudo crear un directorio temporal"

	# --- Las descargas (FR-102) ---

	archivo=kitlegal_${so}_${arq}.tar.gz
	url=$base/$descarga

	curl -fsSL -o "$tmp/$archivo" "$url/$archivo" || fallo "no se pudo descargar $url/$archivo"
	curl -fsSL -o "$tmp/checksums.txt" "$url/checksums.txt" || fallo "no se pudo descargar $url/checksums.txt"

	# --- La huella: la línea cuyo segundo campo es exactamente el archivo (FR-102) ---

	lineas=$(awk -v archivo="$archivo" '$2 == archivo { n++; h = $1 } END { print n + 0, h }' "$tmp/checksums.txt") ||
		fallo "no se pudo leer checksums.txt"
	esperada=${lineas#* }
	case ${lineas%% *} in
	1) ;;
	0) fallo "$archivo no figura en checksums.txt" ;;
	*) fallo "$archivo figura ${lineas%% *} veces en checksums.txt" ;;
	esac

	calculada=$(huella "$tmp/$archivo") || fallo "no se pudo calcular la huella SHA-256 de $archivo"
	calculada=${calculada%% *}
	if [ "$calculada" != "$esperada" ]; then
		fallo "la huella SHA-256 de $archivo ($calculada) no es la de checksums.txt ($esperada)"
	fi

	# --- Solo el miembro kitlegal ---

	mkdir "$tmp/extraido" || fallo "no se pudo crear un directorio temporal"
	tar -xzf "$tmp/$archivo" -C "$tmp/extraido" kitlegal || fallo "no se pudo extraer kitlegal de $archivo"
	binario=$tmp/extraido/kitlegal
	if [ -h "$binario" ] || [ ! -f "$binario" ]; then
		fallo "kitlegal no es un fichero regular en $archivo"
	fi

	# --- La instalación: una copia en el directorio, renombrada encima (FR-103) ---

	mkdir -p "$ruta" || fallo "no se pudo crear el directorio de instalación $dir"
	if [ -d "$ruta/kitlegal" ]; then
		fallo "$dir/kitlegal es un directorio"
	fi
	provisional=$(mktemp "$ruta/.kitlegal.XXXXXX") || fallo "no se pudo crear un temporal en $dir"
	cp "$binario" "$provisional" || fallo "no se pudo copiar kitlegal en $dir"
	chmod 755 "$provisional" || fallo "no se pudo dar permiso de ejecución a kitlegal en $dir"
	mv -f "$provisional" "$ruta/kitlegal" || fallo "no se pudo sustituir $dir/kitlegal"
	provisional=

	# --- Lo que sigue (FR-105, FR-106) ---

	printf 'kitlegal instalado en %s (%s).\n' "$dir/kitlegal" "$que"

	case ":${PATH:-}:" in
	*":$dir:"*) ;;
	*)
		citar "$dir"
		printf '%s no está en el PATH. Para añadirlo en esta sesión, ejecuta la línea siguiente (y, para que dure, añádela al fichero de arranque de tu shell):\n' "$dir"
		printf 'export PATH=%s:"$PATH"\n' "$citado"
		;;
	esac

	printf 'Después, en la carpeta de cada proyecto, instala las skills con:\n'
	printf 'kitlegal skills install\n'
}

{ main "$@"; }
