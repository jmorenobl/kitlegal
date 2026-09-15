# Prueba de red de H5 (quickstart §12.2, SC-012)

Intento 1 de T030, 2026-09-15, propuesta de cambio [#27](https://github.com/jmorenobl/kitlegal/pull/27), cabeza
`6d68c21be163d2f860eefbe1c7e1871f33ba701c`. **Resultado: la ejecución se detuvo antes de la primera sesión**, en el
paso «Retirar Python del runner», con código 100 de `apt-get purge`. No hay informe, ni `sin_python`, ni sesiones: la
prueba de red descubre un defecto del job y T030 se detiene sin marcarse (`gates/tarea-T030.md`; arreglo en T033).

Enlace a la ejecución: <https://github.com/jmorenobl/kitlegal/actions/runs/34922606273> (`databaseId` 34922606273).

## 1. Prerrequisitos (§12.1)

`rtk proxy gh secret list`:

```text
CLAUDE_CODE_OAUTH_TOKEN	2026-09-14T09:17:03Z
CODECOV_TOKEN	2026-09-10T21:18:02Z
```

`rtk proxy gh label list --search evals`:

```text
evals	Ejecuta el job de evals de skills sobre la propuesta de cambio	#1D76DB
evals-prueba-de-red	Ejecuta el job de evals con la prueba de red (SC-012 de H5)	#B60205
```

`gh pr view --json number,headRefOid,labels`:

```json
{"headRefOid":"6d68c21be163d2f860eefbe1c7e1871f33ba701c","labels":[],"number":27}
```

Todo presente: el secreto y las dos etiquetas.

## 2. Órdenes de §12.2

**Primera** (la etiqueta no estaba puesta, así que no se quitó nada):

```text
la etiqueta evals-prueba-de-red no está puesta
```

**Segunda**, `gh pr edit --add-label evals-prueba-de-red`: `https://github.com/jmorenobl/kitlegal/pull/27`.

**Tercera**, a la primera ya con la ejecución:

```text
etiqueta puesta: 2026-09-15T02:48:51Z
{"evals":{"conclusion":"","createdAt":"2026-09-15T02:48:53Z","databaseId":34922606273,"headSha":"6d68c21be163d2f860eefbe1c7e1871f33ba701c","status":"in_progress","url":"https://github.com/jmorenobl/kitlegal/actions/runs/34922606273","workflowName":"evals"},"posteriores_a_la_etiqueta":[{"createdAt":"2026-09-15T02:48:53Z","databaseId":34922606273,"workflowName":"evals"}]}
```

**Cuarta**, `gh run watch 34922606273 --exit-status` (lanzada en segundo plano y esperada hasta el final; último
bloque):

```text
X h5-skill-boe-legislacion evals jmorenobl/kitlegal#27 · 34922606273
Triggered via pull_request about 1 minute ago

JOBS
X evals in 1m33s (ID 104233818438)
  ✓ Set up job
  ✓ Obtener el código del commit evaluado
  ✓ Instalar Go y restaurar la caché
  ✓ Instalar strace y Claude Code
  ✓ Instalar kitlegal y las skills como las deja make install
  X Retirar Python del runner
  - Ejecutar las evals
  - Post Instalar Go y restaurar la caché
  ✓ Post Obtener el código del commit evaluado
  ✓ Complete job

ANNOTATIONS
X Process completed with exit code 100.
evals: .github#109

código 1
```

**Quinta** (informe entre marcas): termina con código 1 **sin imprimir nada**. **Sexta** (salida de la retirada de
Python entre marcas): código 1 **sin imprimir nada**. Las dos leen una ejecución terminada, así que falta alguna marca:
el registro tiene `--- inicio de la retirada de Python ---`, pero el paso se detuvo antes de escribir la de fin, y el
job no llegó a `scripts/evals.sh`, que escribe las del informe.

**Orden de `--log-failed`** (la de tras el bloque, porque la quinta y la sexta fallan con la ejecución terminada),
salida entera tal cual:

```text
evals	Retirar Python del runner	﻿2026-09-15T02:50:26.0952605Z ##[group]Run # Las opciones las fija el propio paso, sin depender de con cuáles lo invoque la plataforma: una orden que falla
evals	Retirar Python del runner	2026-09-15T02:50:26.0954057Z ^[[36;1m# Las opciones las fija el propio paso, sin depender de con cuáles lo invoque la plataforma: una orden que falla^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0954903Z ^[[36;1m# fuera de una condición —la consulta, la purga, la búsqueda o un borrado— lo detiene con su código, y una variable^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0955486Z ^[[36;1m# sin valor, también (research.md D17 y V56).^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0955845Z ^[[36;1mset -euo pipefail^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0956360Z ^[[36;1m# Sin Python accesible para la sesión (FR-073, FR-081; research.md D17). Ninguna orden descarta su salida de error.^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0957070Z ^[[36;1m# scripts/evals.sh repite la búsqueda antes de la primera sesión (contrato del job §3.1).^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0957758Z ^[[36;1m# Las marcas se componen al ejecutar: el texto del paso, si el registro lo reproduce, no las contiene.^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0958295Z ^[[36;1mmarca="retirada de Python"^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0958584Z ^[[36;1mecho "--- inicio de la $marca ---"^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0958863Z ^[[36;1m^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0959263Z ^[[36;1m# Lo que el job usa después de este paso: la retirada no puede llevárselo (research.md D22, S7).^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0959775Z ^[[36;1musados=("$GITHUB_WORKSPACE" "$HOME/.claude")^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0960196Z ^[[36;1mkitlegal=$(readlink -e "$GITHUB_WORKSPACE/bin/instalado/kitlegal")^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0960799Z ^[[36;1musados+=("$kitlegal")^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0961179Z ^[[36;1mfor orden in bash sudo find rm timeout strace claude node go make git; do^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0961981Z ^[[36;1m  if ! ruta=$(command -v "$orden"); then echo "falta $orden" >&2; exit 1; fi^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0962468Z ^[[36;1m  real=$(readlink -e "$ruta")^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0962755Z ^[[36;1m  usados+=("$ruta" "$real")^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0963003Z ^[[36;1mdone^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0963215Z ^[[36;1mcontiene_algo_usado() {^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0963472Z ^[[36;1m  local u^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0963685Z ^[[36;1m  for u in "${usados[@]}"; do^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0963997Z ^[[36;1m    case "$u" in "$1" | "${1%/}"/*) return 0 ;; esac^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0964303Z ^[[36;1m  done^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0964500Z ^[[36;1m  return 1^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0964700Z ^[[36;1m}^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0964886Z ^[[36;1m^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0965363Z ^[[36;1m# 1. Paquetes. Se lista sin patrón y se filtra aquí: con patrones, dpkg-query da también los paquetes que solo conoce^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0966161Z ^[[36;1m# por las relaciones de otros (estado un) y termina con 1 si alguno no casa con nada, así que separar ese caso de un^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0966939Z ^[[36;1m# fallo real obligaría a interpretar su salida de error (research.md V57). Se purga todo paquete con ficheros en^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0967696Z ^[[36;1m# disco, instalado o a medias: solo se salta el que no está (n) o solo conserva su configuración (c). Uno a medias^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0968399Z ^[[36;1m# que no se purgara, apt lo configuraría al purgar los demás, con sus guiones de instalación.^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0968984Z ^[[36;1minstalados=$(dpkg-query -W -f='${db:Status-Abbrev} ${binary:Package}\n')^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0969385Z ^[[36;1mpaquetes=()^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0969626Z ^[[36;1mwhile read -r estado paquete; do^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0969939Z ^[[36;1m  case "$estado" in ?[nc]*) continue ;; esac^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0970397Z ^[[36;1m  case "$paquete" in python*|libpython*|pypy*|libpypy*) paquetes+=("$paquete") ;; esac^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0970844Z ^[[36;1mdone <<< "$instalados"^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0971113Z ^[[36;1mif [ "${#paquetes[@]}" -gt 0 ]; then^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0971429Z ^[[36;1m  echo "paquetes a purgar: ${paquetes[*]}"^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0972060Z ^[[36;1m  sudo apt-get purge -y --auto-remove "${paquetes[@]}"^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0972404Z ^[[36;1melse^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0972632Z ^[[36;1m  echo "paquetes a purgar: ninguno"^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0973107Z ^[[36;1mfi^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0973295Z ^[[36;1m^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0973746Z ^[[36;1m# 2. Todo el sistema de ficheros salvo /proc y /sys, como root: ejecutables y enlaces python* y pypy*, y bibliotecas^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0974526Z ^[[36;1m# libpython* y libpypy*. Se retira cada uno; si está en <prefijo>/bin y <prefijo>/lib tiene un python* o un pypy*,^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0975198Z ^[[36;1m# se retira la instalación entera, salvo que contenga algo de lo que el job usa.^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0976334Z ^[[36;1mbusqueda=(find / '(' -path /proc -o -path /sys ')' -prune -o '(' '(' -type f -perm /111 '(' -iname 'python*' -o -iname 'pypy*' ')' ')' -o '(' -type l '(' -iname 'python*' -o -iname 'pypy*' ')' ')' -o '(' '(' -type f -o -type l ')' '(' -iname 'libpython*' -o -iname 'libpypy*' ')' ')' ')' -print)^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0977334Z ^[[36;1mecho "búsqueda: ${busqueda[*]}"^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0977647Z ^[[36;1mencontrado=$(sudo "${busqueda[@]}")^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0977951Z ^[[36;1mwhile IFS= read -r ruta; do^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0978246Z ^[[36;1m  if [ -z "$ruta" ]; then continue; fi^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0978634Z ^[[36;1m  if ! sudo test -e "$ruta" && ! sudo test -L "$ruta"; then continue; fi^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0979008Z ^[[36;1m  objetivo=$ruta^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0979262Z ^[[36;1m  directorio=$(dirname "$ruta")^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0979567Z ^[[36;1m  if [ "$(basename "$directorio")" = bin ]; then^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0979911Z ^[[36;1m    prefijo=$(dirname "$directorio")^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0980331Z ^[[36;1m    lib=""^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0980582Z ^[[36;1m    if sudo test -d "$prefijo/lib"; then^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0981140Z ^[[36;1m      lib=$(sudo find -H "$prefijo/lib" -mindepth 1 -maxdepth 1 '(' -iname 'python*' -o -iname 'pypy*' ')' -print -quit)^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0981985Z ^[[36;1m    fi^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0982411Z ^[[36;1m    if [ -n "$lib" ] && ! contiene_algo_usado "$prefijo"; then objetivo=$prefijo; fi^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0982838Z ^[[36;1m  fi^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0983054Z ^[[36;1m  echo "retirado: $objetivo"^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0983329Z ^[[36;1m  sudo rm -rf -- "$objetivo"^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0983599Z ^[[36;1mdone <<< "$encontrado"^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0983833Z ^[[36;1m^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0984146Z ^[[36;1m# 3. No queda nada, y la retirada no se ha llevado nada de lo que el job usa.^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0984551Z ^[[36;1mqueda=$(sudo "${busqueda[@]}")^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0984831Z ^[[36;1mif [ -n "$queda" ]; then^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0985211Z ^[[36;1m  while IFS= read -r ruta; do echo "queda Python: $ruta" >&2; done <<< "$queda"^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.0985602Z ^[[36;1m  exit 1^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.1005778Z ^[[36;1mfi^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.1006178Z ^[[36;1mfor u in "${usados[@]}"; do^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.1006933Z ^[[36;1m  if ! sudo test -e "$u"; then echo "la retirada se llevó algo que el job usa: $u" >&2; exit 1; fi^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.1007629Z ^[[36;1mdone^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.1007886Z ^[[36;1mecho "búsqueda tras retirar: ninguno"^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.1008228Z ^[[36;1mecho "--- fin de la $marca ---"^[[0m
evals	Retirar Python del runner	2026-09-15T02:50:26.1063107Z shell: /usr/bin/bash --noprofile --norc -e -o pipefail {0}
evals	Retirar Python del runner	2026-09-15T02:50:26.1063466Z env:
evals	Retirar Python del runner	2026-09-15T02:50:26.1063700Z   MODELO_DE_EVALS: claude-haiku-4-5-20251001
evals	Retirar Python del runner	2026-09-15T02:50:26.1064013Z   VERSION_DE_CLAUDE_CODE: 2.1.270
evals	Retirar Python del runner	2026-09-15T02:50:26.1064281Z   SKILL_EVALUADA: boe-legislacion
evals	Retirar Python del runner	2026-09-15T02:50:26.1064709Z   COMMIT_EVALUADO: 6d68c21be163d2f860eefbe1c7e1871f33ba701c
evals	Retirar Python del runner	2026-09-15T02:50:26.1065162Z   PRUEBA_DE_RED: true
evals	Retirar Python del runner	2026-09-15T02:50:26.1065392Z   GOTOOLCHAIN: local
evals	Retirar Python del runner	2026-09-15T02:50:26.1065606Z ##[endgroup]
evals	Retirar Python del runner	2026-09-15T02:50:26.1121176Z --- inicio de la retirada de Python ---
evals	Retirar Python del runner	2026-09-15T02:50:26.1892881Z paquetes a purgar: libpython3-dev:amd64 libpython3-stdlib:amd64 libpython3.12-dev:amd64 libpython3.12-minimal:amd64 libpython3.12-stdlib:amd64 libpython3.12t64:amd64 python-apt-common python-babel-localedata python-is-python3 python3 python3-apport python3-apt python3-attr python3-automat python3-babel python3-bcrypt python3-blinker python3-boto3 python3-botocore python3-bpfcc python3-certifi python3-cffi-backend:amd64 python3-chardet python3-click python3-colorama python3-commandnotfound python3-configobj python3-constantly python3-cryptography python3-dateutil python3-dbus python3-debconf python3-debian python3-dev python3-distro python3-distro-info python3-distupgrade python3-gdbm:amd64 python3-gi python3-hamcrest python3-httplib2 python3-hyperlink python3-idna python3-incremental python3-jinja2 python3-jmespath python3-json-pointer python3-jsonpatch python3-jsonschema python3-jwt python3-launchpadlib python3-lazr.restfulclient python3-lazr.uri python3-lldb-18 python3-magic python3-markdown-it python3-markupsafe python3-mdurl python3-minimal python3-netaddr python3-netifaces:amd64 python3-netplan python3-newt:amd64 python3-oauthlib python3-openssl python3-packaging python3-parted python3-passlib python3-pexpect python3-pip python3-pip-whl python3-pkg-resources python3-problem-report python3-ptyprocess python3-pyasn1 python3-pyasn1-modules python3-pygments python3-pyparsing python3-pyrsistent:amd64 python3-requests python3-rich python3-s3transfer python3-serial python3-service-identity python3-setuptools python3-setuptools-whl python3-six python3-software-properties python3-systemd python3-twisted python3-typing-extensions python3-tz python3-update-manager python3-urllib3 python3-venv python3-wadllib python3-wheel python3-yaml python3-zope.interface python3-zstandard python3.12 python3.12-dev python3.12-minimal python3.12-venv
evals	Retirar Python del runner	2026-09-15T02:50:26.2125792Z Reading package lists...
evals	Retirar Python del runner	2026-09-15T02:50:26.3770211Z Building dependency tree...
evals	Retirar Python del runner	2026-09-15T02:50:26.3779083Z Reading state information...
evals	Retirar Python del runner	2026-09-15T02:50:26.4723320Z Some packages could not be installed. This may mean that you have
evals	Retirar Python del runner	2026-09-15T02:50:26.4724315Z requested an impossible situation or if you are using the unstable
evals	Retirar Python del runner	2026-09-15T02:50:26.4725335Z distribution that some required packages have not yet been created
evals	Retirar Python del runner	2026-09-15T02:50:26.4726101Z or been moved out of Incoming.
evals	Retirar Python del runner	2026-09-15T02:50:26.4726801Z The following information may help to resolve the situation:
evals	Retirar Python del runner	2026-09-15T02:50:26.4727316Z 
evals	Retirar Python del runner	2026-09-15T02:50:26.4727694Z The following packages have unmet dependencies:
evals	Retirar Python del runner	2026-09-15T02:50:26.5561926Z  shim-signed : Depends: grub-efi-amd64-signed (>= 1.191~) but it is not going to be installed or
evals	Retirar Python del runner	2026-09-15T02:50:26.5563597Z E: Error, pkgProblemResolver::Resolve generated breaks, this may be caused by held packages.
evals	Retirar Python del runner	2026-09-15T02:50:26.5581323Z                         grub-efi-arm64-signed (>= 1.191~) but it is not installable or
evals	Retirar Python del runner	2026-09-15T02:50:26.5593510Z                         base-files (< 12.3) but 13ubuntu10.5 is to be installed
evals	Retirar Python del runner	2026-09-15T02:50:26.5600638Z                Depends: grub-efi-amd64-signed (>= 1.187.2~) but it is not going to be installed or
evals	Retirar Python del runner	2026-09-15T02:50:26.5616039Z                         grub-efi-arm64-signed (>= 1.187.2~) but it is not installable
evals	Retirar Python del runner	2026-09-15T02:50:26.5617582Z                Depends: grub2-common (>= 2.04-1ubuntu24) but it is not going to be installed
evals	Retirar Python del runner	2026-09-15T02:50:26.5628360Z ##[error]Process completed with exit code 100.
```

Lo que se ve: el filtro eligió **110 paquetes** —las bibliotecas y el intérprete del sistema (`python3`,
`python3-minimal`, `python3.12`, `libpython3.12t64:amd64`…) y bibliotecas de las que dependen paquetes del sistema de la
imagen (`python3-apt`, `python3-debconf`, `python3-netplan`, `python3-commandnotfound`, `python3-software-properties`,
`python3-update-manager`, `python3-distupgrade`…)—, y `apt-get purge -y --auto-remove` no encuentra una solución: retirar
esos paquetes arrastra a los que dependen de ellos hasta la cadena de arranque (`shim-signed` exige
`grub-efi-amd64-signed`, que exige `grub2-common`, que ya no se instalaría), y el resolvedor termina con «generated
breaks» y código 100 **antes de cambiar nada**. El paso no llega a su búsqueda ni a ninguna línea `retirado:`.

**Séptima orden** (quitar la etiqueta, al terminar):

```text
https://github.com/jmorenobl/kitlegal/pull/27
la etiqueta evals-prueba-de-red no está puesta
```

## 3. Supuestos de research D22

| Supuesto | Qué muestra esta ejecución | Estado |
|---|---|---|
| **S12** (identificar la ejecución) | (1) La API de eventos devuelve **un** evento `labeled` de la etiqueta: `2026-09-15T02:48:51Z	evals-prueba-de-red	jmorenobl` (`gh api --paginate 'repos/{owner}/{repo}/issues/27/events?per_page=100' --jq '.[] \| select(.event == "labeled") \| "\(.created_at)\t\(.label.name)\t\(.actor.login)"'`). (2) `created_at` (`2026-09-15T02:48:51Z`) y `createdAt` (`2026-09-15T02:48:53Z`) tienen el mismo formato ISO 8601 UTC con `Z`. (3) La ejecución se creó 2 s después del evento. (5) No hay ejecuciones de `evals` anteriores en la rama: la única anterior es la de `ci` (34922178932, antes de la etiqueta), así que la leída no es de otro intento; «quitar y volver a poner» no se ejerció porque la etiqueta no estaba puesta. (6) `posteriores_a_la_etiqueta` lista solo 34922606273, con `workflowName` `evals`, aunque `evals.yml` solo está en la rama de la propuesta de cambio. (4) No se ejerció: la etiqueta no estaba puesta al empezar, y al final `gh pr view` la listaba, así que se quitó estando puesta | **se cumple** en (1), (2), (3) y (6); (4) y (5), sin ejercer |
| **S2** (lo que trae `ubuntu-24.04`) | `sudo` sin contraseña: los pasos con `sudo apt-get` se ejecutaron sin pedirla. `strace`: `strace is already the newest version (6.8-0ubuntu2).` (la imagen ya lo trae). Claude Code: `added 2 packages in 4s` y `claude --version` → `2.1.270 (Claude Code)`. `dpkg-query -W -f='${db:Status-Abbrev} ${binary:Package}\n'` sin patrón terminó con 0 (con `set -euo pipefail` el paso habría parado ahí) y da el nombre con `:amd64` en los paquetes `Multi-Arch: same` (`libpython3.12t64:amd64`, `python3-cffi-backend:amd64`…). **`apt-get purge -y --auto-remove` no retira lo que elige el filtro: termina con 100.** `sudo -n`, `timeout`, findutils y coreutils: sin ejercer (el paso se detuvo antes) | **difiere** en la purga; lo demás ejercido, se cumple |
| **S7** (retirada de Python) | (1) Lo que trae: 110 paquetes de Python en dpkg, entre ellos dependencias de paquetes del sistema; la caché de herramientas, Miniconda y lo que haya fuera de paquetes no se llegó a buscar. (2) Buscar como root: sin ejercer. (3) **Retirar**: ningún paquete retenido aparece en el mensaje, pero la purga es imposible por las dependencias de los paquetes del sistema (`shim-signed` → `grub-efi-amd64-signed` → `grub2-common`). (4) **No romper el job**: la purga, con `--auto-remove`, tendría que retirar paquetes del sistema de los que depende la cadena de arranque, y `apt` la rechaza | **difiere** en (3) y (4); (1) en parte; (2) sin ejercer |
| **S4** (formato de `strace -ff` en el runner) | Ninguna sesión: no hay trazas ni informe | sin evidencia |
| **S9** (una sesión cabe en 240 s) | Ninguna sesión | sin evidencia |
| S1 (fuera de la lista de esta tarea) | `pull_request` con `types: [labeled]` ejecutó el `evals.yml` de la rama (`COMMIT_EVALUADO: 6d68c21…`, `PRUEBA_DE_RED: true`); que el secreto llegue a la sesión no se ejerció | se cumple en lo ejercido |

`sin_python` del informe: **no hay informe**. La comprobación 3 de `scripts/evals.sh` no llegó a ejecutarse.

## 4. Consecuencia

Con el paso tal como está, el job no llega nunca a la primera sesión en `ubuntu-24.04`: ni la prueba de red ni la
ejecución de cierre (T031) pueden funcionar. Research D17 no contemplaba que los paquetes de Python de la imagen fueran
dependencias de paquetes del sistema. El arreglo va en **T033**, antes de T030 en `tasks.md`, con el diagnóstico y las
alternativas en `gates/tarea-T030.md`. El intento siguiente de T030 repite §12.2 entera sobre la cabeza nueva y
reescribe este fichero.
