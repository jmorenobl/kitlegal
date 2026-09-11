# Verificación en la plataforma · escenario 11 (SC-005 · SC-006)

**Fecha**: 2026-09-11 · **Hito**: H0 · **Tarea**: T014 · **Rama**: `h0-esqueleto-del-repo`
**Remoto**: `git@github.com:jmorenobl/kitlegal.git` (`jmorenobl/kitlegal`)
**Intentos**: 1, 2 y 3 (2026-09-10) — bloqueados: no existía `origin`; el 2 corrigió el fixture del guion ·
**4 (2026-09-11, este)** — el prerrequisito humano quedó resuelto (remoto creado y `CODECOV_TOKEN` dado de
alta) y el escenario 11 **se ejecutó entero**.

## Veredicto

**Ejecutado y en verde.** Las dos propuestas de cambio se abrieron contra `h0-esqueleto-del-repo`, se midió
el flujo en frío y en caliente, se cerraron sin integrar y las dos ramas desechables están borradas en local
y en el remoto. **SC-005 y SC-006 quedan validados.**

## Los cuatro resultados y los dos números

| Qué | Criterio | Resultado |
|---|---|---|
| PR sucia bloqueada por lint | SC-005 | **VERDE** · PR #1, control `ci` en `FAILURE`, y el hallazgo es `forbidigo: 1` sobre `internal/prueba/p.go:7:12` |
| PR limpia en verde | SC-006 | **VERDE** · PR #2, los cuatro controles en `SUCCESS` |
| Subida de cobertura termina bien (`CODECOV_TOKEN`) | research.md D15 | **VERDE** · secreto dado de alta; `codecov` con `--fail-on-error` termina en `Upload queued for processing complete`, paso `outcome=success` |
| `internal/core` no figura como fallo ni como pendiente bloqueante | research.md D15 | **VERDE** · `codecov/project/internal/core` → `SUCCESS` (ni `FAILURE` ni `PENDING`) |
| Duración **en caliente** | < 3 min (SC-006) | **1 min 00 s** (`07:01:17Z` → `07:02:17Z`, intento 2 de la ejecución 34572211635, con caché restaurada) |
| Duración **en frío** | se registra, sin umbral | **2 min 48 s** (`06:58:01Z` → `07:00:49Z`, intento 1 de la misma ejecución, `Cache is not found`) |

El escenario en caliente queda a **un tercio** del techo de SC-006, y ni siquiera el frío lo supera: el
diseño de `ci.yml` (un solo job invocando `make ci`, con la caché de `setup-go` cubriendo GOMODCACHE,
GOCACHE y los binarios de las cuatro herramientas) no necesita partirse en jobs paralelos, que es la
corrección que [research.md D12](../research.md) reservaba para el caso contrario.

## Cómo se ejecutó, y por qué así

**Con dos `git worktree` temporales**, la alternativa que el propio escenario declara equivalente
(`quickstart.md` §11, precondición 3). Bajo el workflow `hito` el árbol de la rama del hito siempre tiene
cambios sin confirmar en `gates/`; con worktrees ese árbol **no cambia en ningún momento** y desaparece de
raíz el riesgo de que el registro acabe confirmado en una rama desechable y se vaya con el `git branch -D`.

```console
$ git worktree add /tmp/kl-forbidigo -b prueba-forbidigo h0-esqueleto-del-repo
$ git worktree add /tmp/kl-limpia    -b prueba-limpia    h0-esqueleto-del-repo
```

Las dos ramas salen **de la rama del hito**, no una de otra —es el argumento de arranque de cada
`worktree add`—, y se comprobó sobre el árbol de la limpia antes de tocar nada:

```console
$ git -C /tmp/kl-limpia log --oneline -1
e308077 feat(H0): T014          # el tip del hito, no el 35f7aa4 de la rama sucia
$ git -C /tmp/kl-limpia status --porcelain -- internal/
                                # vacío: no hereda internal/prueba/p.go
$ ls /tmp/kl-limpia/internal/
ls: /tmp/kl-limpia/internal/: No such file or directory
```

Cada commit preparó **solo su fichero** (`git add internal/prueba/p.go` en la sucia, `git add CHANGELOG.md`
en la limpia); ningún `git add -A`. El commit sucio llevó `--no-verify` —el gancho de pre-commit instalado
en T003 ejecuta el mismo lint que ese fichero infringe a propósito, y sin saltárselo no habría propuesta de
cambio que evaluar—; el limpio **no** lo llevó y el gancho pasó.

Al terminar, el árbol de la rama del hito seguía exactamente como estaba (solo las modificaciones previas de
`.claude/settings.json`, `gates/tarea-actual.json` y `gates/tareas-intentos.json`).

## PR sucia — #1 `prueba-forbidigo` (SC-005)

Borrador contra `h0-esqueleto-del-repo`, commit `35f7aa4`. Ejecución **34571956603**, `event: pull_request`,
conclusión `failure` (`06:54:30Z` → `06:57:04Z`, 2 min 34 s).

Antes de empujar, el lint ya fallaba en local con el hallazgo correcto:

```console
$ make -C /tmp/kl-forbidigo lint
internal/prueba/p.go:7:12: use of `fmt.Println` forbidden because "la salida se emite por el escritor
que recibe la función; en H1 solo internal/render escribe en stdout" (forbidigo)
1 issues:
* forbidigo: 1
make: *** [lint] Error 1
```

Y el veredicto de la plataforma —que es el que SC-005 mide— dijo lo mismo:

```
ci	Ejecutar los controles	##[error]internal/prueba/p.go:7:12: use of `fmt.Println` forbidden … (forbidigo)
ci	Ejecutar los controles	1 issues:
ci	Ejecutar los controles	* forbidigo: 1
ci	Ejecutar los controles	make: *** [Makefile:74: lint] Error 1
ci	Ejecutar los controles	##[error]Process completed with exit code 2.
```

```console
$ gh pr checks prueba-forbidigo --json name,state,bucket
[{"bucket":"fail","name":"ci","state":"FAILURE"}]
```

**No basta con que el lint falle: falló por la regla correcta.** Es justo lo que el intento 2 dejó
corregido en el guion (ver más abajo), y aquí se confirma de punta a punta. Los estados de Codecov no
aparecen en esta PR y es lo esperado: el job muere en `make ci`, antes del paso de subida.

## PR limpia — #2 `prueba-limpia` (SC-006)

Borrador contra `h0-esqueleto-del-repo`, commit `3566ccf` (una línea de comentario al final de
`CHANGELOG.md`). Ejecución **34572211635**, dos intentos:

| Intento | Caché | Ventana | Duración | Conclusión |
|---|---|---|---|---|
| 1 (en frío) | `Cache is not found` | `06:58:01Z` → `07:00:49Z` | **2 min 48 s** | `success` |
| 2 (en caliente) | `Cache hit` + `Cache restored successfully` | `07:01:17Z` → `07:02:17Z` | **1 min 00 s** | `success` |

**El frío salió gratis y es real, no provocado:** este repositorio remoto no tenía ninguna ejecución previa
(`gh run list` vacío) y la PR sucia, al fallar, **no** guardó caché —el paso `Post` de `setup-go` no llega a
ejecutarse cuando el job falla—, así que el primer intento de la PR limpia arrancó de veras sin nada. La
caché se guardó al final de ese intento (`Cache saved with the key: setup-go-Linux-x64-ubuntu24-go-1.26.6-…`)
y el rearranque (`gh run rerun`) la reutilizó (`Cache hit occurred on the primary key …, not saving cache`).
No hizo falta invalidar ninguna clave a mano.

Los cuatro controles:

```console
$ gh pr checks prueba-limpia --json name,state,bucket
[{"bucket":"pass","name":"codecov/project/internal/core","state":"SUCCESS"},
 {"bucket":"pass","name":"codecov/patch","state":"SUCCESS"},
 {"bucket":"pass","name":"codecov/project","state":"SUCCESS"},
 {"bucket":"pass","name":"ci","state":"SUCCESS"}]
```

`codeql` no aparece, y debe ser así: su único disparador es `schedule` semanal, precisamente para no cargar
el presupuesto de 3 minutos de SC-006.

### Cobertura: el secreto y el componente

**`CODECOV_TOKEN` está dado de alta.** La orden que la acción ejecuta lleva el token sustituido por
`<redacted>` —señal de que el secreto existía y GitHub lo enmascaró; si faltara, iría vacío— y la subida,
con `--fail-on-error`, terminó bien:

```
./codecov upload-coverage -t <redacted> --fail-on-error --git-service github --sha 3566ccf… --file ./coverage.out
info -- Found 1 coverage files to report
info -- Your upload is now queued for processing …
info -- Upload queued for processing complete
##[end-action id=__codecov_codecov-action.__run_8;outcome=success;conclusion=success;duration_ms=2220]
```

No hizo falta rebajar `fail_ci_if_error`, y no se rebajó. Perfil subido: `coverage: 83.3% of statements`.

**El componente `internal/core` sale `SUCCESS`**, no `FAILURE` ni pendiente que bloquee. Era la duda de
[research.md D15](../research.md): en H0 su `paths` no casa con ningún fichero, y la pregunta abierta era si
Codecov lo dejaría colgado en `PENDING` —que en una rama con controles obligatorios equivale a un bloqueo
permanente— o lo resolvería. Lo resuelve en verde. **Queda contestada.**

## Dependabot y la directiva `toolchain`: no observable todavía, y ahora se sabe por qué

La pregunta —«¿la actualización automática de dependencias propone también subir la directiva
`toolchain`?»— **sigue sin respuesta, pero el motivo ya no es que falte el remoto**. Es concreto y
verificable:

```console
$ git ls-tree -r --name-only origin/main -- .github/
                    # vacío: `main` no tiene .github/ en absoluto
$ git log --oneline -1 origin/main
7b51ac8 fix(speckit): integración claude por defecto en el workflow hito; ignorar runs/
```

Dependabot lee su configuración **de la rama por defecto**. `.github/dependabot.yml` existe solo en
`h0-esqueleto-del-repo`, así que **hoy Dependabot no está activo en este repositorio**: no hay ciclo semanal
que observar ni propuesta que leer. Se activará cuando H0 se integre en `main`, y es entonces —no antes—
cuando se podrá ver si su propuesta sobre `go.mod` toca solo `go 1.26.0` o también `toolchain go1.26.6`.

Lo que consta en el repositorio y no cambia: `CONTRIBUTING.md` §«Subir el parche de Go (directiva
`toolchain`)» describe la subida como un procedimiento **manual** disparado por un hallazgo de `make vuln`
en la biblioteca estándar. Si al integrarse H0 resultara que Dependabot la propone por su cuenta, habrá que
decidir si ese procedimiento manual se solapa. **No se da por buena ninguna de las dos hipótesis aquí**, y
no es un fallo del hito: es una observación que solo la rama por defecto puede producir.

## Cierre y limpieza

Las dos propuestas **cerradas sin integrar** y las dos ramas borradas, local y remotamente, **antes** de
escribir este registro:

```console
$ gh pr list --state all --json number,state,headRefName
[{"headRefName":"prueba-limpia","number":2,"state":"CLOSED"},
 {"headRefName":"prueba-forbidigo","number":1,"state":"CLOSED"}]

$ git branch -a
* h0-esqueleto-del-repo
  main
  remotes/origin/HEAD -> origin/main
  remotes/origin/h0-esqueleto-del-repo
  remotes/origin/main          # ninguna rama de prueba, ni local ni remota

$ git worktree list
/Users/jorge/Projects/kitlegal  e308077 [h0-esqueleto-del-repo]   # los dos temporales, eliminados
```

`gh pr close --delete-branch` borró la rama remota pero no la local (estaba en uso por un worktree); se
completó con `git worktree remove`, `git branch -D` y `git push origin --delete`, y se verificó con
`git ls-remote --heads origin` que el remoto queda solo con `main` y la rama del hito.

---

## Anexo · la discrepancia que el intento 2 corrigió (y que este intento confirma en la plataforma)

Se conserva porque explica por qué el guion del escenario 11 dice lo que dice.

El fixture que el guion dictaba en su redacción original

```bash
printf 'package prueba\n\nimport "fmt"\n\nfunc P() { fmt.Println("no") }\n' > internal/prueba/p.go
```

**no producía el hallazgo de `forbidigo` que su propio apartado «Esperado» anunciaba.** Producía dos de
`revive`:

```console
$ make lint
internal/prueba/p.go:1:1: package-comments: should have a package comment (revive)
internal/prueba/p.go:5:1: exported: exported function P should have comment or be unexported (revive)
2 issues:
* revive: 2
```

**Causa, aislada hasta el final.** El procesador `uniq-by-line` de golangci-lint, activo por omisión, deja
**un único hallazgo por línea**. El `exported` de `revive` cae en la línea 5 (`func P() { … }`) y el de
`forbidigo` también (columna 12 de esa misma línea): el primero gana y el segundo se descarta. Comprobado
por bisección con configuraciones de un solo uso, todas fuera del árbol versionado:

| Configuración | Resultado |
|---|---|
| Solo `forbidigo`, sin exclusiones | `forbidigo: 1` — la regla casa |
| Solo `forbidigo`, **con** el `path-except: ^internal/` del proyecto | `forbidigo: 1` — la exclusión por ruta no es la causa |
| `default: standard` + `forbidigo` | `forbidigo: 1` |
| `default: standard` + `forbidigo` + **`revive`** | `revive: 2`, `forbidigo` desaparece ← reproducido |
| Lo anterior + `issues.uniq-by-line: false` | `forbidigo: 1` **y** `revive: 2` ← confirmado |

La PR sucia habría quedado bloqueada igualmente —el lint termina en error— pero señalando dos comentarios
ausentes, y FR-014 se habría quedado sin comprobar nunca. De ahí que el guion corregido (a) dé al fixture
comentario de paquete y de función exportada, dejando la escritura prohibida como **único** defecto, (b)
exija `forbidigo: 1` en lugar de un fallo genérico y (c) advierta de que **no basta con que el lint falle:
hay que leer qué regla falló**.

La ejecución de hoy confirma la corrección en la plataforma, no solo en local: el log de la ejecución
34571956603 muestra `1 issues: * forbidigo: 1`, exactamente lo que el «Esperado» anuncia.

También quedó comprobado en su día que el mismo `fmt.Println` bajo `cmd/kitlegal/` **no** falla
(`make lint` → `0 issues`): la acotación por ruta de `.golangci.yml` (`path-except: ^internal/`) hace lo
que dice (FR-014).

## Lo que no se hizo, a propósito

- **No se integró ninguna de las dos propuestas.** Ambas quedaron `CLOSED`, nunca `MERGED`.
- **No se rebajó `fail_ci_if_error`** ni ninguna otra bandera. No hizo falta: el secreto estaba dado de alta.
- **No se modificó `.golangci.yml`.** El defecto que apareció en su día estaba en el fixture del guion.
- **No se creó el repositorio remoto ni se dio de alta el secreto**: ambos prerrequisitos humanos ya estaban
  resueltos al empezar este intento, que es lo que lo desbloqueó.
- **No se inventó ningún número.** Los dos de la tabla salen de `startedAt`/`updatedAt` de los dos intentos
  de la ejecución 34572211635, leídos de la plataforma.

## Estado del árbol

El único cambio que esta tarea aporta es este registro. El árbol de la rama del hito no se tocó en ningún
momento —las dos ramas desechables vivieron en worktrees fuera del repositorio—: al terminar solo constan
las modificaciones previas de `.claude/settings.json`, `gates/tarea-actual.json` y `gates/tareas-intentos.json`,
que son cambios sin confirmar del propio bucle del workflow `hito`. `internal/prueba/` no existe y
`git status --porcelain -- internal/` está vacío.
