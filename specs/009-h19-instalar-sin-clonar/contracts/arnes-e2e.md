# Contrato: el arnés e2e y los guiones de aceptación

Lo que los guiones de aceptación de H19 pueden usar del arnés, y de qué contrato sale cada formato que afirman (§6). La
tarea `[aceptacion]` (T001) los escribe contra **este** contrato —binarios, variables, `arbol` y origen local— y, para
lo que el binario o `install.sh` escriben, contra los contratos del producto que enumera §6, en
`specs/009-h19-instalar-sin-clonar/aceptacion/*.txtar`; quedan congelados; el workflow los activa al final
en `internal/app/testdata/script/` con el prefijo `h19-` (V19) y los ejecuta `TestEntregaDelHito`. Las tareas que
implementan el arnés (`internal/app/e2e_test.go` y `internal/app/ejemplo/kitlegal-e2e/main.go`) lo cumplen tal cual.
Decisión en [../research.md](../research.md) D24.

## 1. Regla de todos los guiones: la precondición en rojo

Cada guion empieza, antes de cualquier otra orden, por:

```text
# Precondición: el binario registra el applet skills (FR-010).
exec kitlegal --help
stdout '^  skills +\S'
```

Hoy falla por esa aserción («no match for …»), que es lo que el rojo-primero exige (`scripts/workflow/aceptacion.sh`,
V19): ningún guion puede fallar antes por `exec` («unexpected command failure»), por una orden desconocida o por uso.
Detrás de la precondición, el guion puede usar todo lo de este contrato, que existirá cuando se active, y afirmar los
formatos de salida que fijan los contratos del producto, tal como enumera §6.

## 2. Binarios

Todos son el **mismo** binario de e2e (`internal/app/ejemplo/kitlegal-e2e`: el kernel real con los applets de
ejemplo, `boe` sobre la reproducción de sus grabaciones, `territorio` y **`skills`** con las skills empotradas del
árbol), construidos por `TestMain` fuera del repositorio:

| Cómo se nombra en el guion | Versión (`kitlegal version`) | Creador de enlaces |
|---|---|---|
| `kitlegal` (en el `PATH`; `$KITLEGAL_BIN`) | `dev` (binario de desarrollo: sin forma SemVer) | el del sistema |
| `$KITLEGAL_V1_BIN` (ruta absoluta) | `v0.1.0` | el del sistema |
| `$KITLEGAL_V2_BIN` (ruta absoluta) | `v0.2.0` | el del sistema |
| `$KITLEGAL_SIN_ENLACES_BIN` (ruta absoluta) | `v0.1.0` | uno que siempre falla (`Disponible` falso, `Enlazar` con error) |

El creador de enlaces llega al applet en sus dependencias (`app.AppletSkills(app.DependenciasDeSkills)`, research D9):
los tres primeros usan `app.DependenciasDeSkillsDelSistema(version)` tal cual; el cuarto, las mismas con el campo
`Enlazador` sustituido por un tipo del `package main` de e2e que siempre falla, elegido al construir con una variable de
cadena `-X` de ese paquete y nunca por el entorno (research D24).

Para que las órdenes de `doctor` (que invocan `kitlegal`) usen una versión concreta, el guion la pone delante en el
`PATH` con un enlace llamado `kitlegal`:

```text
mkdir bin
symlink bin/kitlegal -> $KITLEGAL_V1_BIN
env PATH=$WORK/bin${:}$PATH
```

El despacho multicall no tiene ningún applet llamado `kitlegal`, así que toma el applet del primer argumento.

## 3. Variables de entorno de cada guion

| Variable | Valor |
|---|---|
| `KITLEGAL_BIN`, `KITLEGAL_V1_BIN`, `KITLEGAL_V2_BIN`, `KITLEGAL_SIN_ENLACES_BIN` | §2 |
| `KITLEGAL_SKILLS` | ruta absoluta de `skills/` del repositorio (solo lectura): `cmp .agents/skills/boe-legislacion/SKILL.md $KITLEGAL_SKILLS/boe-legislacion/SKILL.md` |
| `KITLEGAL_INSTALADOR` | ruta absoluta de `scripts/install.sh` del repositorio; el arnés la exporta sin exigir que exista, porque el arnés llega antes que `install.sh` en el orden de implementación (plan, pasos 8 y 13) y solo la usan los guiones `instalador-` |
| `KITLEGAL_ORIGEN` | ruta absoluta de un origen de release **de solo lectura para los guiones** (§5); para alterarlo, copiarlo antes: `exec cp -R $KITLEGAL_ORIGEN origen` |
| `KITLEGAL_ORIGEN_VERSION` | versión de la release del origen, sin `v` (`0.1.0` en `make ci`; la de `dist/metadata.json` con `KITLEGAL_DIST`) |
| `KITLEGAL_ORIGEN_ARCHIVO` | nombre del archivo de la plataforma que ejecuta el test (`kitlegal_<GOOS>_<GOARCH>.tar.gz`) |
| `KITLEGAL_CACHE_DIR` | `$WORK/cache` (sin cambios) |
| `http_proxy`, `https_proxy`, `HTTP_PROXY`, `HTTPS_PROXY`, `ALL_PROXY` | `http://127.0.0.1:9`; `NO_PROXY` y `no_proxy` vacíos: toda petición HTTP(S) falla |
| `HOME` | `/no-home` por omisión de testscript (V18); el guion que lo necesita lo fija: `env HOME=$WORK/home` y `mkdir $WORK/home`; sin definir: `exec env -u HOME …`; vacío: `env HOME=` |

La reproducción de `boe` (`$WORK/reproduccion/`) se sigue copiando en cada guion, como hasta ahora.

## 4. La orden `arbol`

```text
arbol <directorio>
```

Escribe en la salida estándar de la orden (la que ven `stdout`, `cp stdout` y `cmp stdout`) una línea por cada entrada
**por debajo** de `<directorio>`, recorrido **sin seguir enlaces**, en orden de bytes de la ruta, con la ruta relativa
a `<directorio>` y `/` como separador:

```text
d .agents
d .agents/skills
f .agents/skills/boe-legislacion/SKILL.md 600 <sha256 hexadecimal>
l .claude/skills/boe-legislacion -> ../../.agents/skills/boe-legislacion
p .agents/skills/legal-core/SKILL.md
o <ruta>
```

`d` directorio, `f` fichero regular (permisos en octal y huella del contenido), `l` enlace simbólico (destino literal),
`p` tubería con nombre, `o` cualquier otra entrada. `<directorio>` tiene que existir y ser un directorio real; si no,
la orden hace fallar el guion. No admite `!`. Es lo que compara «el disco byte a byte»:

```text
arbol .
cp stdout antes.txt
! exec kitlegal skills install          # conflicto: exit 1
arbol .
cmp stdout antes.txt
```

## 5. El origen de release local

`TestMain` construye una vez, en un temporal fuera del repositorio, la disposición de las URL de descarga de GitHub
(research D22), para que `install.sh` la lea con `KITLEGAL_INSTALL_URL=file://<copia>`:

```text
<origen>/latest/download/kitlegal_<GOOS>_<GOARCH>.tar.gz
<origen>/latest/download/checksums.txt
<origen>/download/v<KITLEGAL_ORIGEN_VERSION>/kitlegal_<GOOS>_<GOARCH>.tar.gz
<origen>/download/v<KITLEGAL_ORIGEN_VERSION>/checksums.txt
```

- En `make ci`: el archivo lleva en su raíz el binario de versión `v0.1.0` (el de `KITLEGAL_V1_BIN`) con el nombre
  `kitlegal`, y `checksums.txt` tiene, con el formato de goreleaser (`<sha256>  <nombre>`, ordenado por nombre), la línea
  correcta del archivo y una línea señuelo cuyo nombre **contiene** el del archivo
  (`kitlegal_<GOOS>_<GOARCH>.tar.gz.sbom.json`) con otra huella, para que una búsqueda por subcadena o por posición
  falle.
- Con `KITLEGAL_DIST=<ruta de dist/>` (`make snapshot-check`): el archivo de la plataforma y `checksums.txt` son los
  del snapshot, y `KITLEGAL_ORIGEN_VERSION` la `version` de `dist/metadata.json`. En ese modo `TestEntregaDelHito`
  exige, antes de ejecutar ningún guion, que en `internal/app/testdata/script/` haya al menos un guion cuyo nombre
  contenga `instalador-`; si no, falla con el mensaje `KITLEGAL_DIST: ningún guion instalador- que ejecutar`, para que
  `make snapshot-check` no pase en vacío.

Los guiones `instalador-*` llegan a `internal/app/testdata/script/` como `h19-instalador-*` con la activación, **después**
del bucle de tareas. Hasta entonces, las tareas que los ejecutan —la de `install.sh` contra el origen local, y la de la
release y la de CI con `make snapshot-check`— copian un momento cada `aceptacion/instalador-*.txtar` congelado a
`internal/app/testdata/script/zz-<nombre>.txtar`, ejecutan y las retiran antes de `make ci` y del commit, como hace
`aceptacion.sh rojo-primero` (research D28, «Cómo se verifica dentro del run»). El arnés no distingue unas de otras:
casan con `-run '^TestEntregaDelHito$/instalador-'`.

## 6. Los guiones

Los escribe T001 con estos nombres (el prefijo `h19-` lo añade la activación); cada uno nombra en un comentario de
cabecera los FR y SC que cubre. La tabla de cobertura está en [../plan.md](../plan.md), «Aceptación e2e».

`skills-install-local`, `skills-install-hosts`, `skills-ambito-global`, `skills-ambito-dir`,
`skills-conflictos-entradas`, `skills-conflictos-rutas`, `skills-conflictos-dentro`, `skills-dry-run`,
`skills-idempotencia`, `skills-aviso`, `skills-aviso-sin-aviso`, `skills-list-doctor`, `skills-doctor-hallazgos`,
`skills-doctor-copia`, `skills-no-empotrada`, `skills-invocacion`, `instalador-correcto`, `instalador-rechazos`.

Reglas de escritura: cada proyecto de prueba en su carpeta (`mkdir proyecto` y `cd proyecto`) para no mezclarse con
`reproduccion/` y `cache/`; el código de salida exacto con `exec sh -c '…; test $? -eq N'` (como `boe-codigos.txtar`);
el destino de un enlace con `exec readlink <ruta>` y `stdout '\A…\n\z'`; el JSON por regex sobre `stdout` (`--json`);
ninguna orden a la red.

**De dónde sale cada formato que un guion afirma.** Este contrato solo da el arnés. Lo que escriben el binario e
`install.sh` lo fijan el spec y los contratos del producto, y las tareas que los implementan los cumplen tal cual; un
guion afirma un formato de salida solo si está en el spec de forma literal o en uno de estos apartados, y lo copia de
ahí sin inventar otro:

| Lo que el guion afirma | Contrato |
|---|---|
| mensajes de los errores de invocación (`-g y --dir se excluyen`, `--host no se combina con --dir`, `el único host admitido es claude`, `no es ninguna skill de este binario; skills disponibles: …`, `HOME no está definido o está vacío`) y su precedencia | [applet-skills.md](./applet-skills.md) §2 (FR-052) |
| rutas presentadas en `data` y en los mensajes, por ámbito | [applet-skills.md](./applet-skills.md) §3 |
| claves y valores de `data` de `install`, `list` y `doctor` sin hallazgos (sobre de applet calculado incluido) | [applet-skills.md](./applet-skills.md) §4 |
| cabecera del `mensaje` y líneas `<clase>: <ruta>` de conflictos y `<clase>: <ruta>: <orden>` de hallazgos, con sus clases literales y su orden, en `mensaje` y en stderr; la línea única `skills <verbo>: <clase>: <ruta>` de `list` y `doctor`; el mensaje de fallo de escritura | [applet-skills.md](./applet-skills.md) §5 |
| la orden de cada hallazgo, que SC-011 ejecuta con `sh -c` tras extraerla de la línea `<clase>: <ruta>: <orden>` (lo que sigue al segundo `: `) | [applet-skills.md](./applet-skills.md) §5 y §6 |
| líneas de `--dry-run` | [applet-skills.md](./applet-skills.md) §7 |
| códigos de salida | [applet-skills.md](./applet-skills.md) §8 |
| la línea del aviso, con ` -g` si el manifiesto es el global | [aviso.md](./aviso.md) §4 |
| el manifiesto `kitlegal.json` que un guion escribe a mano o lee | [manifiesto.md](./manifiesto.md) §1 y §3 |
| la forma de invocar en los `SKILL.md` empotrados (`kitlegal <applet> <verbo> …`) | [skills-e-invocacion.md](./skills-e-invocacion.md) §2 |
| `install.sh`: el prefijo `install.sh: ` de los errores y lo que nombran (versión o «la última versión», y lo que falló), la línea `export PATH='<dir>':"$PATH"` y la última línea `kitlegal skills install` | [release.md](./release.md) §7 |

Lo que ningún contrato fija de forma literal (p. ej. el texto de un error del sistema en un fallo de escritura) no se
afirma más allá de lo que el contrato dice que el mensaje contiene.
