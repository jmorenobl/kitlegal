# Evidencia de plataforma de H4 (T037)

Esta evidencia recoge la publicación de la rama del hito, la apertura de la propuesta de cambio y la lectura de los
estados de la integración continua y de Codecov. Cubre el punto 1 de la Definition of Done leído **en remoto**, el §6
del roadmap (ritual por hito) y ADR 0007.

- **Rama**: `h4-applet-boe-puerto`
- **Commit publicado**: `40efb064a5a3f3e6b1ca2e3eae8d751f85d5d878` (`feat(H4): T036`), 43 commits por delante de
  `main` (`2c889c4`, igual en local y en `origin` tras `git fetch`)
- **Propuesta de cambio**: [#24](https://github.com/jmorenobl/kitlegal/pull/24) → `main`, abierta, no borrador,
  `MERGEABLE`
- **Fecha**: 2026-09-13
- **Intento**: 1 de 3. **T037 queda sin marcar** (causa y arreglo en [`tarea-T037.md`](./tarea-T037.md))

**Lo que esta tarea no hace** (ADR 0007): no fusiona, no empuja a `main`, no fuerza, no borra ramas y no crea etiquetas
ni releases. La fusión es siempre humana.

## Resumen

| Comprobación | Resultado |
|---|---|
| Rama publicada, con el gancho `pre-push` ejecutado | ✅ `* [new branch]`, `guardia-push` en verde |
| Propuesta abierta con [`pr-h4.md`](./pr-h4.md) como cuerpo | ✅ #24 |
| Trabajo `ci` | ❌ run `34781187264`, `failure` en 2 m 13 s: `TestDependenciasDelBinario` en el ejecutor linux (§3) |
| `codecov/project`, `codecov/project/internal/core`, `codecov/project/internal/cli`, `codecov/patch` | ⚠️ **no emitidos**: la subida del perfil no se ejecuta tras el fallo de `make ci` (§5); ni verde vacío ni configuración |
| Causa del rojo | Un test cuyo veredicto depende de la plataforma que lo ejecuta, no un cambio de superficie (§4) |
| S9 (`boe-cache-rapida` < 200 ms en el ejecutor) | Primera medida a favor: `internal/app` en `ok` bajo `-race` (§6) |
| Arreglo | T038, comprobado sobre un clon con `make ci` en verde; T037 lo publica en su intento 2 (§7) |
| Sin fusión, sin push a `main`, sin `--force`, sin etiquetas | ✅ |

---

## Aviso de método

El envoltorio de terminal de esta máquina reescribe la salida de `go test`, `git status --porcelain`, `git diff` y
`gh pr checks`. Todo lo que sigue se leyó con el paso directo (`rtk proxy`) y se copia literal, con dos excepciones
señaladas: `git push` y `gh pr create`, que se ejecutaron tal como los escribe la tarea.

---

## 1. Publicación de la rama

Antes de empujar, `git ls-remote --heads origin h4-applet-boe-puerto` no devolvió ninguna línea, y `origin/main`
coincidía con `main` (`2c889c4`, que es ancestro de `HEAD`).

```
$ git push -u origin h4-applet-boe-puerto
╭─────────────────────────────────────╮
│ 🥊 lefthook v1.13.6  hook: pre-push │
╰─────────────────────────────────────╯
┃  guardia-push ❯

  ────────────────────────────────────
summary: (done in 0.01 seconds)
✔️ guardia-push (0.01 seconds)
remote:
remote: Create a pull request for 'h4-applet-boe-puerto' on GitHub by visiting:
remote:      https://github.com/jmorenobl/kitlegal/pull/new/h4-applet-boe-puerto
remote:
To github.com:jmorenobl/kitlegal.git
 * [new branch]      h4-applet-boe-puerto -> h4-applet-boe-puerto
branch 'h4-applet-boe-puerto' set up to track 'origin/h4-applet-boe-puerto'.
```

```
$ rtk proxy git rev-parse HEAD origin/h4-applet-boe-puerto
40efb064a5a3f3e6b1ca2e3eae8d751f85d5d878
40efb064a5a3f3e6b1ca2e3eae8d751f85d5d878
$ rtk proxy git rev-list --count main..HEAD
43
```

El gancho `pre-push` dejó pasar la referencia: es una rama de hito, no `main`, no va forzada, no es un borrado ni una
etiqueta.

## 2. Apertura de la propuesta de cambio

```
$ gh pr view h4-applet-boe-puerto
no pull requests found for branch "h4-applet-boe-puerto"

$ gh pr create --base main --head h4-applet-boe-puerto --title "feat(H4): applet boe, puerto de boe.py" \
    --body-file specs/005-h4-applet-boe-puerto/gates/pr-h4.md
ok created #24 https://github.com/jmorenobl/kitlegal/pull/24
```

(La segunda salida es la forma abreviada del envoltorio; la dirección es la de la propuesta.)

```
$ rtk proxy gh pr view 24 --json number,state,isDraft,baseRefName,headRefName,headRefOid,mergeable,statusCheckRollup,url
{"baseRefName":"main","headRefName":"h4-applet-boe-puerto","headRefOid":"40efb064a5a3f3e6b1ca2e3eae8d751f85d5d878","isDraft":false,"mergeable":"MERGEABLE","number":24,"state":"OPEN","statusCheckRollup":[{"__typename":"CheckRun","completedAt":"2026-09-13T20:36:39Z","conclusion":"FAILURE","detailsUrl":"https://github.com/jmorenobl/kitlegal/actions/runs/34781187264/job/103788376613","name":"ci","startedAt":"2026-09-13T20:34:26Z","status":"COMPLETED","workflowName":"ci"}],"url":"https://github.com/jmorenobl/kitlegal/pull/24"}
```

El cuerpo es [`pr-h4.md`](./pr-h4.md) tal como lo dejó T036: la plantilla del ritual y la sección «Dependencias» con
la justificación de §V. Tras leer los estados se corrigió y la propuesta se actualizó con el mismo fichero (§7).

## 3. Integración continua

```
$ rtk proxy gh run list --branch h4-applet-boe-puerto
completed	failure	feat(H4): applet boe, puerto de boe.py	ci	h4-applet-boe-puerto	pull_request	34781187264	2m17s	2026-09-13T20:34:23Z
```

Solo el flujo `ci` se dispara con una propuesta de cambio: CodeQL y `nightly` son programados.

```
$ rtk proxy gh run view 34781187264 --json jobs --jq '.jobs[] | "JOB \(.name)\t\(.conclusion)\t\(.startedAt)\t\(.completedAt)", (.steps[] | "  \(.number)\t\(.name)\t\(.conclusion)\t\(.startedAt)\t\(.completedAt)")'
JOB ci	failure	2026-09-13T20:34:26Z	2026-09-13T20:36:39Z
  1	Set up job	success	2026-09-13T20:34:27Z	2026-09-13T20:34:30Z
  2	Obtener el código	success	2026-09-13T20:34:30Z	2026-09-13T20:34:32Z
  3	Instalar Go y restaurar la caché	success	2026-09-13T20:34:32Z	2026-09-13T20:35:17Z
  4	Ejecutar los controles	failure	2026-09-13T20:35:17Z	2026-09-13T20:36:36Z
  5	Publicar el perfil de cobertura	skipped	2026-09-13T20:36:36Z	2026-09-13T20:36:36Z
  9	Post Instalar Go y restaurar la caché	skipped	2026-09-13T20:36:36Z	2026-09-13T20:36:36Z
  10	Post Obtener el código	success	2026-09-13T20:36:37Z	2026-09-13T20:36:37Z
  11	Complete job	success	2026-09-13T20:36:37Z	2026-09-13T20:36:37Z
```

La caché de `setup-go` se restauró: H4 no cambia `go.sum` frente a `main`.

```
ci	Instalar Go y restaurar la caché	2026-09-13T20:35:17.7836467Z Cache restored successfully
ci	Instalar Go y restaurar la caché	2026-09-13T20:35:17.8133767Z Cache restored from key: setup-go-Linux-x64-ubuntu24-go-1.27.1-faad8ae2390159a627abd187a613292da7f4afabaf9cffc8224f62e16a4222ae
```

Registro del paso que falla (`rtk proxy gh run view 34781187264 --log-failed`). Se omiten las líneas de las dos listas
que coinciden; la lista A son los 18 declarados y la B, los 16 medidos:

```
ci	Ejecutar los controles	2026-09-13T20:35:18.2255069Z go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
ci	Ejecutar los controles	2026-09-13T20:35:20.0096787Z go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
ci	Ejecutar los controles	2026-09-13T20:35:44.0903885Z 0 issues.
ci	Ejecutar los controles	2026-09-13T20:35:44.1293302Z go test -race -shuffle=on -coverprofile=coverage.out ./...
ci	Ejecutar los controles	2026-09-13T20:35:48.1556969Z ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	1.013s	coverage: 0.0% of statements
ci	Ejecutar los controles	2026-09-13T20:35:48.1557964Z -test.shuffle 1789331747542698468
ci	Ejecutar los controles	2026-09-13T20:35:48.1558467Z --- FAIL: TestDependenciasDelBinario (0.13s)
ci	Ejecutar los controles	2026-09-13T20:35:48.1573190Z     arch_test.go:209: 
ci	Ejecutar los controles	2026-09-13T20:35:48.1602334Z         	Error:      	elements differ
ci	Ejecutar los controles	2026-09-13T20:35:48.1651662Z         	            	extra elements in list A:
ci	Ejecutar los controles	2026-09-13T20:35:48.1665531Z         	            	([]interface {}) (len=2) {
ci	Ejecutar los controles	2026-09-13T20:35:48.1665531Z         	            	 (string) (len=26) "github.com/mattn/go-isatty",
ci	Ejecutar los controles	2026-09-13T20:35:48.1681948Z         	            	 (string) (len=30) "github.com/ncruces/go-strftime"
ci	Ejecutar los controles	2026-09-13T20:35:48.1751604Z         	            	listA:
ci	Ejecutar los controles	2026-09-13T20:35:48.1751604Z         	            	([]string) (len=18) {
[…]
ci	Ejecutar los controles	2026-09-13T20:35:48.1936884Z         	            	listB:
ci	Ejecutar los controles	2026-09-13T20:35:48.1937337Z         	            	([]string) (len=16) {
[…]
ci	Ejecutar los controles	2026-09-13T20:35:48.1975007Z         	Messages:   	el binario distribuido enlaza un módulo que no está declarado y justificado (FR-060, constitución §V): justifícalo en plan.md y en la propuesta de cambio antes de añadirlo
ci	Ejecutar los controles	2026-09-13T20:35:48.1976576Z FAIL	github.com/jmorenobl/kitlegal/internal	0.285s
ci	Ejecutar los controles	2026-09-13T20:35:54.1152147Z ok  	github.com/jmorenobl/kitlegal/internal/app	5.098s	coverage: 93.7% of statements
ci	Ejecutar los controles	2026-09-13T20:35:58.5265201Z ok  	github.com/jmorenobl/kitlegal/internal/cache	8.002s	coverage: 90.5% of statements
ci	Ejecutar los controles	2026-09-13T20:35:58.5291619Z ok  	github.com/jmorenobl/kitlegal/internal/cli	1.349s	coverage: 98.6% of statements
ci	Ejecutar los controles	2026-09-13T20:35:58.5311779Z ok  	github.com/jmorenobl/kitlegal/internal/core/schema	1.017s	coverage: 90.1% of statements
ci	Ejecutar los controles	2026-09-13T20:36:36.1028127Z ok  	github.com/jmorenobl/kitlegal/internal/httpx	34.082s	coverage: 97.2% of statements
ci	Ejecutar los controles	2026-09-13T20:36:36.1029609Z ok  	github.com/jmorenobl/kitlegal/internal/render	1.014s	coverage: 95.8% of statements
ci	Ejecutar los controles	2026-09-13T20:36:36.1038198Z ok  	github.com/jmorenobl/kitlegal/internal/source/boe	5.458s	coverage: 99.4% of statements
ci	Ejecutar los controles	2026-09-13T20:36:36.1249216Z make: *** [Makefile:67: test] Error 1
ci	Ejecutar los controles	2026-09-13T20:36:36.1264920Z ##[error]Process completed with exit code 2.
```

Formato y lint pasaron en remoto (`0 issues.`), y todos los paquetes salvo `internal` quedaron en `ok`, también
`internal/source/boe` (99,4 %) y `internal/app` (93,7 %). `make` se detuvo en `test`, así que `test-integration`,
`vuln`, `schema-check`, `secrets`, `mod-verify` y `mod-tidy-check` no llegaron a ejecutarse en esta pasada.

## 4. Causa: el veredicto del test dependía de la plataforma que lo ejecuta

`TestDependenciasDelBinario` (T026) lanza `go list -deps -f '{{if .Module}}{{.Module.Path}}{{end}}' ./cmd/kitlegal`
con el entorno del proceso, así que mide el binario de la plataforma **en la que corre el test**. Los módulos que
enlaza el binario cambian con la plataforma. Medido en local, sin red y sin cgo, como se distribuye (ADR 0002):

```
$ rtk proxy sh -c 'for p in darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64; do printf "%s: " "$p"; CGO_ENABLED=0 GOOS=${p%/*} GOARCH=${p#*/} go list -deps -f "{{if .Module}}{{.Module.Path}}{{end}}" ./cmd/kitlegal | sort -u | grep -v -x -e "" -e github.com/jmorenobl/kitlegal | tr "\n" " "; echo; done'
```

| Plataforma | Módulos | Sin |
|---|---|---|
| darwin/amd64, darwin/arm64 | 18 | — |
| linux/amd64, linux/arm64 | 16 | `github.com/mattn/go-isatty`, `github.com/ncruces/go-strftime` |
| windows/amd64, windows/arm64 | 17 | `github.com/google/uuid` |

La razón está en las restricciones de compilación de `modernc.org/libc@v1.75.6`, en la caché de módulos:

```
## google/uuid
libc_darwin.go  (sin restricción)
libc_musl.go  //go:build linux && (amd64 || arm64 || loong64 || ppc64le || s390x || riscv64 || 386 || arm)
libc_unix.go  //go:build unix && !(linux && (amd64 || arm64 || loong64 || ppc64le || s390x || riscv64 || 386 || arm))
[… freebsd, illumos, netbsd, openbsd; ningún fichero de windows]
## mattn/go-isatty
libc.go  //go:build !linux || mips64le
## ncruces/go-strftime
libc_unix.go  //go:build unix && !(linux && (amd64 || arm64 || loong64 || ppc64le || s390x || riscv64 || 386 || arm))
libc_windows.go  (sin restricción)
```

- **La superficie declarada es correcta.** `modulosDelBinario` son exactamente los 18 de la unión de las seis
  plataformas, y `go.mod`, `go.sum` y `codecov.yml` no cambian frente a `main`. El defecto es del test, no de la lista.
- **Por qué no salió antes.** Todos los `make ci` del hito, incluido el escenario 16 de T036, se ejecutaron en
  darwin/arm64, donde el binario enlaza los 18. El ejecutor `ubuntu-latest` es la primera ejecución en otra plataforma.
- **Reproducido en local.** Un binario de test nativo cuyo `go list` hijo hereda la plataforma de linux, sobre un clon
  desechable de `40efb06`, falla igual que el run:

```
$ rtk proxy go -C /tmp/kitlegal-t037/copia test -count=1 -exec 'env GOOS=linux GOARCH=amd64 CGO_ENABLED=0' -run '^TestDependenciasDelBinario$' ./internal/
--- FAIL: TestDependenciasDelBinario (0.17s)
    arch_test.go:209:
        	Error:      	elements differ
        	            	extra elements in list A:
        	            	([]interface {}) (len=2) {
        	            	 (string) (len=26) "github.com/mattn/go-isatty",
        	            	 (string) (len=30) "github.com/ncruces/go-strftime"
        	            	}
[…]
FAIL	github.com/jmorenobl/kitlegal/internal	0.532s
```

## 5. Estados de la propuesta: Codecov no emitió ninguno

```
$ rtk proxy gh pr checks h4-applet-boe-puerto
ci	fail	2m13s	https://github.com/jmorenobl/kitlegal/actions/runs/34781187264/job/103788376613

$ rtk proxy gh api repos/jmorenobl/kitlegal/commits/40efb064a5a3f3e6b1ca2e3eae8d751f85d5d878/check-runs --jq '.check_runs[] | "\(.name)\t\(.status)\t\(.conclusion)\t\(.started_at)\t\(.completed_at)\t\(.output.title)"'
ci	completed	failure	2026-09-13T20:34:26Z	2026-09-13T20:36:39Z	null
```

`gh pr checks --watch` terminó con código 1, y hay un único estado. **Faltan los cuatro de Codecov, y la causa no es
la configuración**:

- «Publicar el perfil de cobertura» (`codecov/codecov-action`) salió `skipped`, porque GitHub Actions no ejecuta los
  pasos siguientes a uno que ha fallado. Sin perfil subido, Codecov no evalúa ni publica nada.
- No es un verde sobre cero ficheros, ni falta un estado en `codecov.yml`: `project`, `patch` y los componentes
  `internal_core` e `internal_cli` siguen declarados. `ci.yml` sube los dos perfiles con `fail_ci_if_error: true`.
- Los estados solo se pueden leer en una ejecución en la que `make ci` pase. Es lo que hace el intento 2 (§8). Como H4
  no crea el árbol de ningún componente, `internal_core` e `internal_cli` medirán de verdad contra la base de `main`.

## 6. S9: primera medida de `boe-cache-rapida` en el ejecutor

S9 (research D20) supone que las diez invocaciones con caché caliente bajan de 200 ms en el ejecutor de la plataforma,
con la suite en paralelo y `-race`. En este run, `make test` (`go test -race -shuffle=on … ./...`) dejó
`internal/app` en `ok` a los 5,098 s. Ese paquete ejecuta `TestEntregaDelHito`, cuyos guiones incluyen
`boe-cache-rapida`:

- `internal/app/e2e_test.go` no tiene restricción de compilación ni `t.Skip` (`grep -c t.Skip` → `0`).
- `boe-cache-rapida.txtar` no tiene ninguna condición de salto (`grep -E 'skip|\[short\]|\[!'` no encuentra nada).

**Primera medida a favor de S9**, con una reserva: sin `-v`, el registro no nombra el subtest. La ejecución en verde del
intento 2 la repite, y la evidencia final la tomará de allí.

## 7. Arreglo y cambios de este intento

T037 no puede arreglar el test: `internal/arch_test.go` está fuera de sus rutas. Tampoco serviría redelimitarla, porque
el workflow commitea la tarea después del intento, y el push de este intento publicaría `HEAD` sin el arreglo. Por eso:

- **T038**, en `tasks.md` justo antes de T037. `siguiente_tarea` elige la primera tarea sin marcar. T038 declara solo
  `internal/arch_test.go` (comprobado con la extracción de rutas del workflow).
  - **Qué hace**: declara las seis plataformas de distribución, mide el cierre de cada una con `GOOS`, `GOARCH` y
    `CGO_ENABLED=0` y compara la unión con `modulosDelBinario`. La lista no cambia, no se rebaja nada y no se añade
    ninguna supresión.
  - **Comprobado sobre un clon desechable**:
    - el rojo reproducido (§4);
    - `--- PASS` en nativo y con el `-exec` de linux/amd64 y de windows/arm64;
    - `golangci-lint run ./internal/` → `0 issues.`;
    - las dos sondas negativas (quitar `go-isatty` y declarar un módulo ficticio) en `FAIL`, con las plataformas de cada
      módulo en el mensaje;
    - `make -C … ci` → `ci: todos los controles en verde`.

  El diff está en [`tarea-T037.md`](./tarea-T037.md).
- **`pr-h4.md`**: la sección «Dependencias» explica que los módulos enlazados dependen de la plataforma, y la tabla dice
  a qué plataformas llegan `uuid`, `go-isatty` y `go-strftime`. «Pendientes» recoge este rojo, T038 y la primera
  medida de S9. La propuesta se actualizó con ese fichero (`gh pr edit 24 --body-file …`).

## 8. Qué queda para el intento 2 de T037

Cuando T038 esté commiteada:

1. Volver a empujar la rama, en avance rápido y sin forzar.
2. `gh pr view` encontrará la #24 y no se creará otra.
3. Esperar `ci` y leer los cuatro estados de Codecov por el título de su check-run: `project` ≥ 70 %,
   `internal/core` ≥ 85 %, `internal/cli` ≥ 90 %, y `patch` con objetivo `auto`.
4. Completar esta evidencia y «Pendientes» de `pr-h4.md`, actualizar el cuerpo de la propuesta y ejecutar `make ci`.
   Solo entonces se marca.

## `make ci` final de este intento

Se ejecutó en primer plano con todos los ficheros del intento ya escritos salvo esta sección, así que `gitleaks` los
recorrió. T037 no toca código: el árbol es `40efb06` más los cambios del directorio del feature.

```
$ rtk proxy sh -c 'make ci > /tmp/kitlegal-t037/repo-make-ci.log 2>&1; echo "código $?"; grep -E "^(ok|FAIL|---|ci:|\?)|issues|DATA RACE|vulnerabilities|leaks|all modules|schema-check|tidy" /tmp/kitlegal-t037/repo-make-ci.log'
código 0
0 issues.
ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	1.752s	coverage: 0.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal	1.590s	coverage: [no statements]
ok  	github.com/jmorenobl/kitlegal/internal/app	5.079s	coverage: 93.7% of statements
ok  	github.com/jmorenobl/kitlegal/internal/cache	5.063s	coverage: 90.7% of statements
ok  	github.com/jmorenobl/kitlegal/internal/cli	2.493s	coverage: 98.6% of statements
?   	github.com/jmorenobl/kitlegal/internal/core	[no test files]
ok  	github.com/jmorenobl/kitlegal/internal/core/schema	2.860s	coverage: 90.1% of statements
ok  	github.com/jmorenobl/kitlegal/internal/httpx	14.508s	coverage: 97.2% of statements
ok  	github.com/jmorenobl/kitlegal/internal/render	3.388s	coverage: 95.8% of statements
ok  	github.com/jmorenobl/kitlegal/internal/source/boe	3.730s	coverage: 99.4% of statements
ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	(cached)	coverage: 0.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal	(cached)	coverage: [no statements]
ok  	github.com/jmorenobl/kitlegal/internal/app	(cached)	coverage: 93.7% of statements
ok  	github.com/jmorenobl/kitlegal/internal/cache	(cached)	coverage: 96.2% of statements
ok  	github.com/jmorenobl/kitlegal/internal/cli	(cached)	coverage: 98.6% of statements
?   	github.com/jmorenobl/kitlegal/internal/core	[no test files]
ok  	github.com/jmorenobl/kitlegal/internal/core/schema	(cached)	coverage: 90.1% of statements
ok  	github.com/jmorenobl/kitlegal/internal/httpx	(cached)	coverage: 97.2% of statements
ok  	github.com/jmorenobl/kitlegal/internal/render	(cached)	coverage: 95.8% of statements
ok  	github.com/jmorenobl/kitlegal/internal/source/boe	(cached)	coverage: 99.4% of statements
No vulnerabilities found.
ok  	github.com/jmorenobl/kitlegal/internal/app	0.863s
go tool -modfile=tools/gitleaks/go.mod gitleaks dir . --redact --no-banner
10:55PM INF no leaks found
all modules verified
== tools/gitleaks
all modules verified
all modules verified
all modules verified
all modules verified
go mod tidy -diff
ci: todos los controles en verde
```

(Se han quitado los códigos de color de la línea `INF` de `gitleaks`.)

- **`internal` en `ok` aquí y en `FAIL` en remoto no se contradicen: es exactamente el defecto de §4.** Esta máquina es
  darwin/arm64, donde el binario enlaza los 18 módulos. Este verde local no sustituye al remoto, y por eso T037 no se
  marca.
- **`test-integration` sale `(cached)`**, y es legítimo: desde `40efb06` no ha cambiado ningún fichero Go ni material de
  test.
- **Veredicto del intento 1**: la verificación determinista está en verde, pero el criterio de la tarea (el estado de
  la integración continua y de Codecov en remoto) no se cumple. **T037 queda `[ ]`**, con la causa en
  [`tarea-T037.md`](./tarea-T037.md) y el arreglo en T038.
