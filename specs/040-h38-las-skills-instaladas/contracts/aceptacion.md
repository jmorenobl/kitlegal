# Contrato: la aceptación e2e de H38

La primera tarea, `[aceptacion]`, escribe cuatro guiones testscript en `specs/040-h38-las-skills-instaladas/aceptacion/`,
uno por historia con comportamiento del binario, y quedan congelados (ADR 0018). El workflow los activa como
`internal/app/testdata/script/h38-<nombre>.txtar` (`scripts/workflow/aceptacion.sh activar`), donde los ejecuta
`TestEntregaDelHito` en `make ci`. US5 es documentación (`README.md`, `CHANGELOG.md`) y no tiene guion; US4.5 es el
unitario `TestReinstalacionFalloAMitad` (FR-019; research.md D7).

## §1 · Lo común a los cuatro

- **Binarios**: los del arnés de H19, `$KITLEGAL_V1_BIN` (`v0.1.0`) y `$KITLEGAL_V2_BIN` (`v0.2.0`), con las mismas
  skills empotradas (`specs/009-h19-instalar-sin-clonar/contracts/arnes-e2e.md`). Ninguno abre red: los proxies del
  arnés están cerrados (SC-005).
- **`HOME` por caso**: `env HOME=$WORK/<caso>` y `mkdir`; con hosts, `mkdir $HOME/.claude $HOME/.gemini/config` antes
  de instalar, para que `install -g` los enlace.
- **Caché por caso**: `env KITLEGAL_CACHE_DIR=$WORK/cache-<caso>`, que no existe hasta que algo la crea.
- **La consulta**: `territorio resolver Leganés --json --no-graph`, que no escribe en la caché (research.md D10). El
  exit 4 es `boe articulo BOE-A-2015-10565 a99 --offline` dentro de `exec sh -c '…; test $? -eq 4'`.
- **Referencias**: al principio de cada guion, cada orden del guion con `HOME` vacío y sin skills: su salida estándar
  (`cp stdout`), su salida de error (`cp stderr`) y su código (`test $? -eq N`).
- **«Exactamente la línea»** (`\A…\n\z`): donde la referencia deja vacía la salida de error, `stderr '\A<línea>\n\z'`.
- **«Una línea de aviso más»**: donde la referencia ya escribe en la salida de error (`--dry-run`, un error sin
  `--json`), `una-linea-mas.sh <ref> <obtenida> <patrón>`, un guion de shell del propio archivo txtar: la obtenida
  tiene exactamente una línea que empieza por `aviso: `, casa con el patrón, y sin ella es byte a byte la referencia.
- **«Nada escrito»** (FR-013): `arbol $HOME` antes y después, comparados con `cmp`, y `! exists $KITLEGAL_CACHE_DIR`.
- **«Reinstala»** (FR-005): `arbol $HOME` igual al de un `HOME` gemelo con la misma preparación sobre el que se
  ejecuta `$KITLEGAL_V2_BIN skills install -g` directo; `! grep 'v0\.1\.0' $HOME/.agents/skills/kitlegal.json`; y
  `! exists $KITLEGAL_CACHE_DIR/skills-global.lock`.

Las líneas, con las versiones del arnés:

```text
FR-006: aviso: kitlegal v0\.2\.0 ha actualizado las skills globales, que eran de v0\.1\.0; se aplican en la próxima conversación
FR-009: aviso: kitlegal v0\.2\.0 se actualizarían las skills globales, que eran de v0\.1\.0; se aplican en la próxima conversación
hoy:    aviso: las skills instaladas son de kitlegal v0\.1\.0 y este binario es kitlegal v0\.2\.0; ejecuta: kitlegal skills install -g
```

## §2 · `aceptacion/reinstala.txtar` (US1)

Cubre: US1.1 a US1.8 · FR-001, FR-003, FR-004, FR-005, FR-006, FR-008, FR-012, FR-013 · SC-001, SC-003.

| Caso | Preparación | Se comprueba |
|---|---|---|
| primera y segunda consulta (US1.1-1.3) | hosts; `V1 skills install -g` | la primera: salida estándar y código de la referencia, `\A<FR-006>\n\z`, reinstala; la segunda: `! stderr .`, el árbol igual; `V2 skills doctor -g --json` sin hallazgos |
| falta un fichero (US1.4) | hosts; `V1 install -g`; `rm` de un fichero de `references/` | `\A<FR-006>\n\z`, reinstala y el fichero vuelve |
| una skill en otra versión (US1.5) | hosts; `V1 install -g`; `V2 skills install -g plazos-y-territorio` | `\A<FR-006>\n\z` (nombra `v0.1.0`), reinstala |
| un nombre retirado intacto (US1.6) | sin hosts; `V1 install -g`; `mv` de `normativa` a `boe-legislacion` y `sed 's/"normativa/"boe-legislacion/g'` en `kitlegal.json` | `\A<FR-006>\n\z`; `exists …/normativa/SKILL.md`; `! exists …/boe-legislacion`; `! grep boe-legislacion …/kitlegal.json`; el árbol el del gemelo sin hosts |
| exit 4 (US1.7) | hosts; `V1 install -g` | `! stdout .`, código 4, `una-linea-mas.sh` con FR-006, reinstala |
| multicall (US1.8) | hosts; `V1 install -g`; `symlink $WORK/bin/territorio -> $KITLEGAL_V2_BIN` | `exec $WORK/bin/territorio resolver Leganés --json --no-graph`: salida de la referencia, `\A<FR-006>\n\z`, reinstala |

## §3 · `aceptacion/tocadas.txtar` (US2)

Cubre: US2.1 a US2.4 · FR-003, FR-004, FR-005, FR-007, FR-013 · SC-002, SC-003, SC-006.

Empieza por un **control** —hosts, `V1 install -g` sin tocar nada, la consulta reinstala con `\A<FR-006>\n\z`— que
prueba que en cada caso lo que impide reinstalar es lo tocado y que hace fallar hoy el guion por una aserción.

| Caso | Preparación (cada uno en su `HOME`, desde `V1 install -g`) | Se comprueba |
|---|---|---|
| fichero editado (US2.1) | `echo … >> …/plazos-y-territorio/SKILL.md` | `\A<hoy>\n\z`; nada escrito; `grep '"version": "v0\.1\.0"'` en el manifiesto |
| fichero ajeno (US2.2) | `olvida.sh` retira del manifiesto la declaración de `plazos-y-territorio/references/jerarquia_normativa.md` | `\A<hoy>\n\z`; nada escrito |
| entrada de host ajena (US2.2) | hosts; `rm $HOME/.claude/skills/normativa` y `symlink` a `$WORK/fuera` | `\A<hoy>\n\z`; nada escrito |
| retirada «sin retirar» (US2.2) | sin hosts; el nombre retirado de §2 y `echo … >> …/boe-legislacion/SKILL.md` | `\A<hoy>\n\z`; nada escrito |
| binario más viejo (US2.3) | `V2 install -g`; la consulta con `$KITLEGAL_V1_BIN` | `\Aaviso: las skills instaladas son de kitlegal v0\.2\.0 y este binario es kitlegal v0\.1\.0; ejecuta: kitlegal skills install -g\n\z`; nada escrito |
| metadatos de construcción (US2.4) | `V2 install -g`; `sed 's/"v0\.2\.0"/"0.2.0+abc"/g'` en el manifiesto | `\Aaviso: las skills instaladas son de kitlegal 0\.2\.0\+abc y este binario es kitlegal v0\.2\.0; ejecuta: kitlegal skills install -g\n\z`; nada escrito |

En todos, la salida estándar y el código son los de la referencia (SC-003).

## §4 · `aceptacion/excluidas.txtar` (US3)

Cubre: US3.1 a US3.5 y el local inservible de «Edge Cases» · FR-001, FR-002, FR-003 a, FR-009, FR-013, FR-014,
FR-015 · SC-003, SC-006.

Empieza por el mismo **control** que §3, en su propio `HOME`. Después, otro `HOME` con hosts y `V1 install -g`, sin
tocar, sobre el que se lanzan todas las invocaciones excluidas, y tras cada una «nada escrito»:

| Caso | Invocación de `$KITLEGAL_V2_BIN` | Se comprueba |
|---|---|---|
| `mcp serve` (US3.3) | `mcp llamadas.txt $KITLEGAL_V2_BIN mcp serve --no-graph`, con `herramientas` y `territorio_resolver {"consulta":"Leganés"}` | `grep -count=1 '^aviso: ' mcp.err`; esa línea es la de hoy; la expectativa de `h21-mcp-protocolo.txtar` no cambia |
| `--describe` (US3.4) | `territorio resolver Leganés --describe --no-graph` | la salida estándar de la referencia; `\A<hoy>\n\z` |
| análisis fallido (US3.4) | `boe articulo`, código 2 | `! stdout .`; `una-linea-mas.sh` con la de hoy |
| ayuda (US3.4, FR-002) | `territorio resolver --help`, `territorio --help`, `version` | la salida estándar de la referencia; `! stderr .` |
| `--dry-run` (US3.5) | `territorio resolver Leganés --dry-run --no-graph` | `! stdout .`; `una-linea-mas.sh` con FR-009 |

`kitlegal skills …` (FR-002) no va en este guion: su salida estándar es lo que hay instalado, así que no tiene
referencia sin skills. Que el applet `skills` no busca el aviso lo vigilan las filas `skills-*` de
`TestAvisoDelKernel` (`internal/app/aviso_test.go:191-195`, `avisa: false`), donde el kernel no da ninguna línea del
avisador, y los guiones de H19.

Y el ámbito local, cada caso en su directorio de proyecto y con el `HOME` de las excluidas, cuyo global sigue en la v1
porque ninguna de ellas reinstala:

| Caso | Preparación | Se comprueba |
|---|---|---|
| local viejo (US3.1) | `V1 skills install` en el proyecto | `\Aaviso: las skills instaladas son de kitlegal v0\.1\.0 y este binario es kitlegal v0\.2\.0; ejecuta: kitlegal skills install\n\z`; `arbol` del proyecto y de `HOME` sin cambios |
| local al día con el global viejo (US3.2) | `V2 skills install` en otro proyecto | `! stderr .`; `arbol` de `HOME` sin cambios |
| local inservible con el global viejo (FR-003 a, «Edge Cases») | otro proyecto cuyo `./.agents` es un fichero del propio archivo del guion, como `agents-fichero` en `h19-skills-aviso-sin-aviso.txtar` | la salida estándar de la referencia; `! stderr .`; `arbol` de `HOME` sin cambios; `! exists $KITLEGAL_CACHE_DIR` |

## §5 · `aceptacion/candado.txtar` (US4)

Cubre: US4.1 a US4.4 y US4.6 · FR-010, FR-011, FR-013 · SC-003, SC-004.

| Caso | Preparación (hosts; `V1 install -g`) | Se comprueba |
|---|---|---|
| candado reciente (US4.1) | `mkdir $KITLEGAL_CACHE_DIR`; `cp candado.txt $KITLEGAL_CACHE_DIR/skills-global.lock` | `\A<hoy>\n\z`; el árbol sin cambios; el candado sigue y `cmp` con `candado.txt` |
| candado huérfano (US4.2) | como el anterior y `exec touch -t 202001010000 $KITLEGAL_CACHE_DIR/skills-global.lock` | `\A<FR-006>\n\z`; reinstala; `! exists …/skills-global.lock` |
| sin directorio de caché (US4.3) | `KITLEGAL_CACHE_DIR` que no existe | `\A<FR-006>\n\z`; reinstala; `exists $KITLEGAL_CACHE_DIR`; `! exists …/skills-global.lock` |
| sin permiso de escritura (US4.4) | `exec chmod -R a-w $HOME/.agents/skills` | `\A<hoy>\n\z`; con `exec chmod -R u+w $HOME/.agents/skills`, el árbol igual al de antes; `! exists …/skills-global.lock`; la consulta siguiente, `\A<FR-006>\n\z` y reinstala |
| dos a la vez (US4.6) | — | `exec … &a&` y `exec … &b&`, `wait a` y `wait b` con la salida de cada una en un fichero: una sola línea de FR-006 entre las dos, la otra la de hoy o vacía, la salida estándar de cada una la de la referencia, y el árbol el de una reinstalación |

Los permisos se devuelven antes de terminar el caso, como en `h19-instalador-rechazos.txtar`, para que testscript
pueda borrar `$WORK`.

## §6 · Los e2e de H19 que cambian (FR-018)

Los cambia una tarea `[datos]` que declara las seis rutas. Cambian las líneas de los controles globales, que ahora
reinstalan, y su comentario. Y como cada control deja el global en la v2, tras el último control de cada guion al que
siguen casos sobre el mismo `HOME`, `exec $KITLEGAL_V1_BIN skills install -g` lo devuelve a la v1: esos casos siguen
corriendo con el global de otra versión que dice su comentario, que es lo que los hace distinguir un aviso que mirase
el global, y ninguna de sus expectativas cambia. Las expectativas locales y `h21-mcp-protocolo.txtar` no cambian.

| Fichero activo y su copia congelada | Línea (la de hoy) | Pasa a |
|---|---|---|
| `internal/app/testdata/script/h19-skills-aviso.txtar` y `specs/009-h19-instalar-sin-clonar/aceptacion/skills-aviso.txtar` | 119 | `stderr '\A<FR-006>\n\z'`, y tras ella `exec $KITLEGAL_V1_BIN skills install -g`, antes de los casos de 121-132 (manifiesto local viejo y al día con el global de otra versión, `HOME=$WORK/home-global`) |
| `internal/app/testdata/script/h19-skills-aviso-sin-aviso.txtar` y `specs/009-h19-instalar-sin-clonar/aceptacion/skills-aviso-sin-aviso.txtar` | 152, 156 y 160 | `stderr '\A<FR-006>\n\z'`, con `exec $KITLEGAL_V1_BIN skills install -g` antes de los controles de 155 y 159 y tras el de 160, antes de los casos de 162-189 (`agents-fichero`, `skills-colgando`, `skills-enlace` y `local-ilegible`, `HOME=$WORK/home-v1`) |
| `internal/app/testdata/script/h19-skills-conflictos-dentro.txtar` y `specs/009-h19-instalar-sin-clonar/aceptacion/skills-conflictos-dentro.txtar` | 203 | `stderr '\A<FR-006>\n\z'`, y tras la 204 (`cp stdout $WORK/territorio.json`) `exec $KITLEGAL_V1_BIN skills install -g`, antes del caso (h) de 206-230 (`HOME=$WORK/home-v1`) |

Con esas devoluciones, los casos de un local inservible con un global de otra versión de los dos últimos guiones
(`./.agents` fichero, `./.agents/skills` enlace colgando o al global, `kitlegal.json` ilegible o enlazado) siguen
saliendo en rojo si el aviso mirase el global: daría la línea de FR-006 y fallaría su `! stderr .`.

Comprobado en esta sesión: las líneas están en esas posiciones en los seis ficheros (`grep -n 'install -g'`; las de
los casos que siguen a cada control, con `grep -n 'env HOME\|V[12]_BIN territorio\|! stderr'` en la copia activa y
la congelada; `diff` entre las dos solo da líneas que esto no toca, como los códigos de `install`, `list` y `doctor`
en 216-222 del caso (h)), y
ningún otro guion de `internal/app/testdata/script/` reinstalaría: `grep -l 'skills install.* -g'` da nueve guiones, y
en los otros seis ningún `skills install … -g` es de `$KITLEGAL_V1_BIN`, `$KITLEGAL_V2_BIN` ni
`$KITLEGAL_SIN_ENLACES_BIN` (`grep`, sin coincidencias): instalan el global con `kitlegal`, el binario de desarrollo,
cuya versión no tiene forma SemVer; y su única consulta con un binario SemVer (`skills-salida-legible.txtar:116`)
compara un manifiesto local.
