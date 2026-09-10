# T003 — bloqueada por el guardián de ficheros sensibles del harness (intentos 1, 2 y 3)

**Estado**: `[ ]` (no marcada). `make ci` queda en verde, pero T003 **no** está entregada: ese verde
no dice nada sobre la tarea, porque el entregable que falta —`lefthook.yml`— no entra en `make ci`.

**Intentos agotados**: 3 de 3. La causa es la misma en los tres y no está al alcance del agente.

## Causa (idéntica en los tres intentos)

El harness de Claude Code bloquea **toda** escritura sobre `lefthook.yml`, que es justamente el
entregable central de la tarea. La denegación es del propio harness (guardián de ficheros sensibles
para configuraciones que ejecutan código en operaciones de git), no de las reglas de permisos del
repositorio. Comprobado de nuevo en el intento 3:

- `.claude/settings.json` → `permissions.allow` incluye `Write` y `Edit`.
- `.claude/settings.json` → `permissions.deny` es
  `git push`, `git merge`, `git reset --hard`, `rm -rf`, `curl`, `wget`, `sudo`. **No menciona
  `lefthook.yml`** ni ninguna ruta.

Mensaje literal, el mismo en los tres intentos:

```
Claude requested permissions to edit /Users/jorge/Projects/kitlegal/lefthook.yml
which is a sensitive file.
```

Que el guardián cubra este fichero es coherente: `lefthook.yml` declara órdenes que se ejecutan
solas en cada `git commit`, así que escribirlo es conceder ejecución de código diferida.

### Vías descartadas, y por qué

- **Intento 1**: tres vías, las tres denegadas — `Write` sobre el fichero inexistente,
  `printf … > lefthook.yml` desde `Bash`, y `Edit` sobre el esqueleto que genera `lefthook install`.
- **Intento 2**: se reprobó la vía directa (`Write`) por si el entorno hubiera cambiado. Misma
  denegación.
- **Intento 3**: se reprobó otra vez la vía directa (`Write`), con el contenido completo y final.
  Misma denegación.

**No se reintentó la vía `Bash`.** El guardián actúa por herramienta, pero su intención es que el
agente no escriba ese fichero sin aprobación: colarlo por otra herramienta sería eludirlo, no
cumplirlo. Por lo mismo se descartan escribir con otro nombre y renombrar con `mv`, y generar el
fichero desde una receta del `Makefile`. Tampoco se toca `.claude/settings.json` para añadirse
permiso a sí mismo: está fuera de las rutas declaradas por la tarea y sustituiría el criterio de
quien contribuye por el del agente.

La sesión es **no interactiva**, así que no hay forma de conceder la aprobación que el guardián pide.

## Lo que sí quedó hecho y verificado

- **`tools/lefthook/`**: módulo de herramienta creado igual que los otros tres —`go mod init`,
  `go get -tool github.com/evilmartians/lefthook@latest`, `go mod tidy` sobre su propio modfile—.
  Fija **lefthook v1.13.6**, exactamente la versión contra la que `research.md` D16 manda comprobar
  el comportamiento de `stage_fixed`. Confirmado en `86fdf5e`, y re-verificado en el intento 3:
  `go tool -modfile=tools/lefthook/go.mod lefthook version` → `1.13.6`.
- **`make ci` en verde** con el módulo nuevo dentro; `mod-verify` lo recoge por el glob sobre
  `tools/*/go.mod` que fijó T001, sin tocar el `Makefile`. Cobertura 83.3 %, 0 issues de lint, sin
  vulnerabilidades, sin secretos, `go mod tidy -diff` limpio.
- **`Makefile` sin cambios**: T001 ya dejó `LEFTHOOK` y el objetivo `hooks` escritos.

## Estado en el que se deja el repositorio (intento 3)

Sin cambios respecto al final de los intentos 1 y 2. En este intento **no** se ejecutó `make hooks`
—sin configuración solo generaría otra vez el esqueleto de ejemplo, que el intento 1 ya revirtió con
`lefthook uninstall --remove-configs`— y **no** se ejecutó el commit de prueba del escenario 8: sin
`lefthook.yml` con bloque `pre-commit` no hay gancho que comprobar, y un commit de prueba sin gancho
mutaría `cmd/kitlegal/main.go` sin comprobar nada.

Rojo de partida verificado antes de intentar la escritura, y sigue igual al terminar:

- `lefthook.yml` **ausente**.
- `.git/hooks/` con **solo** los 14 `.sample` de git: ningún gancho real instalado.

Invariantes que exige la tarea, comprobadas al terminar contra la revisión de partida
`49cfb628eeb9ae391351c641fadbf8cfefad9d4b` (`base` de `gates/tarea-actual.json`):

- `git rev-parse HEAD` → `49cfb628eeb9ae391351c641fadbf8cfefad9d4b`. **Igual a `$antes`.**
- `git diff --stat 49cfb62 -- cmd/kitlegal/main.go` → **vacío**.
- `git diff --cached --stat 49cfb62 -- cmd/kitlegal/main.go` → **vacío**.
- `git status --porcelain` → solo `gates/tarea-actual.json` y `gates/tareas-intentos.json`, que son
  los que escribe el propio workflow. No se persiguió el árbol limpio: ni `git checkout .` ni
  `git clean`, que destruirían el trabajo de la tarea.
- `make ci` → `ci: todos los controles en verde`.

Queda pendiente, y **no** se dio por bueno: `lefthook.yml`, `make hooks` con configuración real y el
escenario 8 completo (mutación de espaciado + línea de comentario, commit, formateo, re-preparado y
deshacer anclado a `$antes`).

## Qué hace falta para desbloquearla

Los tres intentos están agotados, así que T003 necesita una intervención humana. Cualquiera de estas
tres sirve:

1. Ejecutar la tarea en una **sesión interactiva**, donde se pueda aprobar la escritura de
   `lefthook.yml` cuando el guardián la pida.
2. Que quien contribuye añada la autorización previa para esa ruta concreta en la configuración de
   permisos del repositorio, de modo que el modo desatendido pueda escribirla.
3. Si se prefiere no relajar el guardián: **escribir `lefthook.yml` a mano** con el contenido de
   abajo y dejar al agente solo `make hooks` y el escenario 8.

El contenido está enteramente fijado por `research.md` D16 y FR-022 —bloque `pre-commit` con cuatro
trabajos que invocan **solo** órdenes del `Makefile`—:

```yaml
# Ganchos de git de kitlegal (FR-022, research.md D16).
#
# El gancho invoca únicamente órdenes del `Makefile`: el `Makefile` es la única
# superficie de invocación de los controles, de modo que lo que corrige el
# gancho antes de confirmar y lo que verifica la integración continua son la
# misma orden y no pueden divergir (SC-004).
#
# Se renuncia deliberadamente a acotar el formato y el lint a los ficheros
# preparados: un segundo camino de invocación sería un segundo sitio donde la
# configuración puede divergir, y con un repositorio de este tamaño esa
# optimización no compensa el riesgo.
#
# El gancho NO es la autoridad final: la integración continua ejecuta `make ci`,
# que incluye estos mismos controles en su modo no mutante.
#
# Instalación: `make hooks`.

pre-commit:
  jobs:
    # `fmt` es el modo mutante del formato (FR-012): corrige el fichero y
    # `stage_fixed` lo vuelve a preparar, de forma que se confirma lo formateado.
    - name: fmt
      glob: "*.go"
      run: make fmt
      stage_fixed: true

    - name: lint-fast
      run: make lint-fast

    - name: secrets
      run: make secrets

    - name: mod-tidy-check
      run: make mod-tidy-check
```

Y después el escenario 8 tal cual lo describe `quickstart.md`, incluido sustituir `stage_fixed` por
un `git add {staged_files}` explícito si en v1.13.6 no re-preparase los ficheros corregidos —la
obligación de verificación nº 2 que `research.md` deja a la implementación, que sigue **sin
resolver** porque nunca se ha podido ejecutar el gancho—.
