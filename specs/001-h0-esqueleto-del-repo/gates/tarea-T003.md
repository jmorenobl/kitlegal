# T003 — intento 1: bloqueada por el guardián de ficheros sensibles del harness

**Estado**: `[ ]` (no marcada). `make ci` queda en verde, pero T003 **no** está entregada.

## Causa

El harness de Claude Code bloquea **toda** escritura sobre `lefthook.yml`, que es justamente el
entregable central de la tarea. La denegación es del propio harness (guardián de ficheros sensibles
para configuraciones que ejecutan código en operaciones de git), no de las reglas de permisos del
repositorio: `.claude/settings.json` permite `Write`, `Edit` y `Bash(make:*)`, y su lista `deny` no
menciona `lefthook.yml`.

Se intentaron tres vías, las tres denegadas con el mismo mensaje —
`Claude requested permissions to edit /Users/jorge/Projects/kitlegal/lefthook.yml which is a sensitive file`:

1. Herramienta `Write` sobre el fichero inexistente.
2. `printf … > lefthook.yml` desde `Bash`.
3. Herramienta `Edit` sobre el fichero ya existente (el esqueleto de ejemplo que genera `lefthook install`).

La sesión es **no interactiva**, así que no hay forma de conceder la aprobación que el guardián pide.
No se buscó ningún rodeo (escribir con otro nombre y renombrar con `mv`, por ejemplo): eso vaciaría de
sentido el guardián y sustituiría el criterio de quien contribuye por el del agente.

## Lo que sí quedó hecho y verificado

- **`tools/lefthook/`**: módulo de herramienta creado igual que los otros tres —`go mod init`,
  `go get -tool github.com/evilmartians/lefthook@latest`, `go mod tidy` sobre su propio modfile—.
  Fija **lefthook v1.13.6**, exactamente la versión contra la que `research.md` D16 manda comprobar
  el comportamiento de `stage_fixed`.
- **`make ci` en verde** con el módulo nuevo dentro. `mod-verify` lo recoge por sí solo a través del
  glob sobre `tools/*/go.mod` que fijó T001, sin tocar el `Makefile`:

  ```
  == tools/gitleaks      all modules verified
  == tools/golangci-lint all modules verified
  == tools/govulncheck   all modules verified
  == tools/lefthook      all modules verified
  ```

  Cobertura 83.3 %, 0 issues de lint, sin vulnerabilidades, sin secretos.
- **`Makefile` sin cambios**: T001 ya dejó `LEFTHOOK` y el objetivo `hooks` escritos.

## Estado en el que se deja el repositorio

`make hooks` llegó a ejecutarse (rojo → el módulo faltaba; verde → el módulo ya está) y lefthook,
al no encontrar configuración, **generó él mismo un `lefthook.yml` de ejemplo** —solo comentarios— y
sincronizó un `prepare-commit-msg` en `.git/hooks/`. Ni el esqueleto de ejemplo ni ese gancho
implementan nada de FR-022, y dejar el esqueleto habría hecho que el paso de commit del workflow
confirmara un fichero de ejemplo vacío de contenido bajo una ruta declarada. Se revirtió con
`lefthook uninstall --remove-configs`, que retira el gancho y el esqueleto sin tocar nada más.

El commit de prueba del escenario 8 **nunca llegó a ejecutarse**, porque sin `lefthook.yml` con bloque
`pre-commit` no hay nada que comprobar. Las invariantes que exige la tarea se cumplen de forma trivial
y se verificaron igualmente:

- `git rev-parse HEAD` = `cb610ac46f8cc9544026579d97601fe942182033` (la revisión de partida, `base` de
  `gates/tarea-actual.json`).
- `git diff --stat cb610ac -- cmd/kitlegal/main.go` vacío.

Queda pendiente, y **no** se dio por bueno: `lefthook.yml`, `make hooks` con configuración real y el
escenario 8 completo (mutación de espaciado + línea de comentario, commit, formateo, re-preparado y
deshacer anclado a `$antes`).

## Qué hace falta para el intento siguiente

Una de estas dos, ambas fuera del alcance del agente:

1. Ejecutar el intento en una **sesión interactiva**, donde se pueda aprobar la escritura de
   `lefthook.yml` cuando el guardián la pida.
2. Añadir una autorización previa para esa ruta concreta en la configuración de permisos del
   repositorio, de modo que el modo desatendido pueda escribirla.

Una vez desbloqueado, el contenido a escribir está enteramente fijado por `research.md` D16 y FR-022:
bloque `pre-commit` con cuatro trabajos que invocan **solo** órdenes del `Makefile` —`make fmt` con
`glob: "*.go"` y `stage_fixed: true`, `make lint-fast`, `make secrets` y `make mod-tidy-check`—, y
después el escenario 8 tal cual lo describe `quickstart.md`, incluido el sustituir `stage_fixed` por
un `git add {staged_files}` explícito si en v1.13.6 no re-preparase los ficheros corregidos.
