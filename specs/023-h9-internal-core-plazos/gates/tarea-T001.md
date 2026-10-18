# T001, intento 2 (2026-10-18): la suite está completa; `make ci` en rojo en el árbol de trabajo por un residuo del arnés que la sesión no puede retirar

## Para el intento siguiente, antes de leer nada más

```
ls -d internal/source/boe/testdata/boe.legislacion-consolidada/.claude testdata/evals/boe.legislacion-consolidada/.claude
```

- **Si alguno existe**: no repitas ninguna verificación (ya se hizo dos veces, abajo). Deja la tarea `[ ]`, añade a esta
  nota la fecha y el número del intento y termina. La sesión no puede retirarlos: `rmdir` sobre una ruta `.claude/` pide
  aprobación («sensitive file») y en modo desatendido nadie la da; borrarlos con otra orden (`rm -r`, `find -delete`,
  `git clean`, un programa en Go) sería sortear ese rechazo, y no corresponde.
- **Si no existe ninguno**: `make ci` en primer plano y, en verde, marca T001 `[X]`. Los guiones no se tocan: cumplen la
  tarea tal como están.

## Lo que está hecho y no hay que tocar

Los cinco guiones de `specs/023-h9-internal-core-plazos/aceptacion/` (`plazos-calcular`, `plazos-cobertura`,
`plazos-codigos`, `plazos-herramienta`, `plazos-sin-efectos`), escritos en el intento 1 desde spec.md y
contracts/applet-plazos.md, y su línea `[comportamiento] T001` en `gates/supuestos.md` (la procedencia
`kitlegal.plazos` también en los sobres de error, órdenes y herramienta: `plazos-herramienta` la afirma en `mcp-2.json`).

Verificado en el intento 1 (nota anterior): prototipo de un solo uso bajo `$TMPDIR` con el applet mínimo de contracts y
el calendario mínimo de research D27, los cinco en verde (25 de 25 con `-count=5`) y los cinco mutantes de la tarea
atrapados por su guion, más once mutantes propios; el prototipo se retiró.

Verificado por el verificador del workflow tras el intento 1 (`gates/ci.log`, 07:29): `guardián de diff ok (T001, 11
ficheros)` y `aceptación: 5 guiones en rojo por una aserción; evals nuevas: 0`. Los guiones no han cambiado desde las
07:10 (fecha de modificación).

Verificado en este intento, sin tocar los guiones:

- Cada uno lleva una vez `exec kitlegal --help` seguido de `stdout '^\s+plazos\s'` antes de cualquier otra orden, y
  ninguno contiene `usage: `, `unknown command`, `unexpected command` ni `cannot parse`. Las afirmaciones que fija la tarea
  están (fechas, preceptos, `source` de los festivos, códigos, clases, `plazos_calcular`, `arbol cache`, `--dry-run`,
  `--offline --no-graph`, la frase negativa de Tordesillas).
- **`make ci` en el árbol de trabajo: rojo**, solo en `test`, por una única causa, la misma del intento 1:
  `read ../../internal/source/boe/testdata/boe.legislacion-consolidada/.claude: is a directory` y `el conjunto de
  grabaciones ../../internal/source/boe/testdata/boe.legislacion-consolidada no se puede copiar: .claude no es un fichero
  regular` (`internal/evals`, `internal/skills`; en el intento 1 también `internal/app`, `test-integration` y
  `skills-check`, 83 veces en `gates/ci.log`). `TestInterrumpirLaSesion` cae por «Condition never satisfied» porque su
  sesión no llega a prepararse por esa misma causa.
- **`make ci` en un árbol limpio de `HEAD` (8209889): verde.** `git worktree add --detach` bajo `$TMPDIR` y
  `make -C … ci` desde la raíz del repositorio: `0 issues.`, `no leaks found`, `ci: todos los controles en verde`. Ese
  árbol no llevaba la suite (copiar ficheros fuera del repositorio pide aprobación en esta sesión): `make ci` no la
  ejecuta, y en el árbol de trabajo, que sí la lleva, los únicos rojos son los del residuo. El árbol limpio se retiró con
  `git worktree remove` y `git worktree prune`.

## La causa

`internal/source/boe/testdata/boe.legislacion-consolidada/.claude/.cc-writes/` es un directorio **vacío** que deja el
arnés de Claude Code (la marca de escrituras de su sandbox; el propio sandbox nombra `<directorio>/.claude/.cc-writes`
entre sus rutas). Se creó el 2026-10-18 a las 05:09:32, durante el paso `plan` de este run (spec.json es de las 05:02 y
plan.json de las 06:31). Lo ignora git por `~/.config/git/ignore` (`**/.claude/.cc-writes/`): `git status` sale limpio y
el guardián de diff no lo cuenta, pero los tests que leen o copian el directorio de las grabaciones lo encuentran y, con
razón, no admiten en él nada que no sea un fichero regular. Nada de esta tarea toca esos tests ni ese directorio.

Hay 41 directorios así en el árbol (`find . -path ./.git -prune -o -type d -name .cc-writes -print`), todos vacíos, con
fechas de sesiones interactivas (1, 5, 6 y 16 de octubre) y de los pasos de este run (04:36 `specify`; 05:09, 06:11 y
06:27 `plan`). Hoy solo rompe `make ci` el de las grabaciones de `boe`; el de
`testdata/evals/boe.legislacion-consolidada/.claude` (06:11:19) puede romper la tarea `[datos]` o cualquier test que
liste ese directorio.

Qué los crea, hasta donde se pudo medir aquí: ni la herramienta de escritura (sonda en `gates/sonda-write/`, retirada)
ni la de lectura (`scripts/workflow/` y `.specify/workflows/hito/`, leídos en esta sesión, siguen sin `.claude/`), y
desde las 07:29 no ha aparecido ninguno nuevo. Queda como candidato el directorio de trabajo de una orden de la
herramienta Bash (un `cd` a un subdirectorio), que esta sesión no puede probar porque ese `cd` pide aprobación.

## Lo que hace falta, y solo lo puede hacer una persona

1. **Parar el run** antes del intento 3 (TERM a `scripts/hito.sh`): la sesión de escalada se gastaría contra el mismo
   muro, y las tareas siguientes también caerían en `make ci` hasta el «DAG bloqueado» de tres cuarentenas seguidas.
2. **Retirar los residuos**, sin tocar nada versionado. Lo mínimo:
   ```
   rmdir internal/source/boe/testdata/boe.legislacion-consolidada/.claude/.cc-writes internal/source/boe/testdata/boe.legislacion-consolidada/.claude
   rmdir testdata/evals/boe.legislacion-consolidada/.claude/.cc-writes testdata/evals/boe.legislacion-consolidada/.claude
   ```
   O todos los vacíos de una vez, que `-empty` protege lo que tiene contenido (`.claude/` de la raíz):
   ```
   find . -path ./.git -prune -o -type d -name .cc-writes -path '*/.claude/*' -empty -print -delete
   find . -path ./.git -prune -o -type d -name .claude -empty -print -delete
   ```
3. **Reanudar** con `scripts/hito.sh --resume <run_id>`; `gates/tareas-intentos.json` lleva `T001: 2`, y los dos
   intentos cayeron por el entorno, no por la tarea.
4. **Arreglo de raíz, fuera de la rama del hito** (es proceso: `scripts/workflow/*` y `scripts/hito.sh` los rechaza el
   guardián dentro de una tarea): que el workflow retire esos directorios vacíos antes de medir —en
   `scripts/workflow/verificar.sh` antes de `make ci`, o en `scripts/hito.sh` antes de cada paso—, porque son residuo del
   arnés y no del árbol, y los crean las propias sesiones del run. Una regla de git no basta: el problema está en el
   sistema de ficheros, no en lo que git ve.

## Reanudación (2026-10-18, supervisor)

Los residuos `.claude/.cc-writes` se retiraron y ya no vuelven: el workflow 2.9.1 (#161, cherry-pick `f266c12` en esta rama) los retira al abrir cada sesión, tras cada herramienta y antes de `make ci`. La tarea salió de la cuarentena con su contador a cero. Su trabajo de los intentos anteriores está en `gates/cuarentena/T001.patch` (vacío en T001: sus guiones siguen en `aceptacion/`); devuélvelo al árbol con `git apply` y verifica con `make ci` en primer plano.

## Intento 1 tras la reanudación (2026-10-18): verde

Sin residuos `.claude/` en las dos rutas de grabaciones. `cuarentena/T001.patch` vacío: nada que aplicar; los guiones no
se tocaron. «Rojo primero» a mano (copias `zz-rojo-plazos-*` en `internal/app/testdata/script/`, retiradas después): los
cinco fallan con `no match for ^\s+plazos\s found in stdout`, la aserción de su precondición. `make ci` en primer plano:
`0 issues.`, `no leaks found`, `ci: todos los controles en verde`. T001 marcada `[X]`. El prototipo no se repitió en
esta sesión: su resultado es el que registra el intento 1, arriba.
