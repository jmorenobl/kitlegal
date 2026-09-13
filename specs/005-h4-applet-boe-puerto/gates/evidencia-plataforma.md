# Evidencia de plataforma de H4 (T037)

Esta evidencia recoge la publicación de la rama del hito, la apertura de la propuesta de cambio y la lectura de los
estados de la integración continua y de Codecov. Cubre el punto 1 de la Definition of Done leído **en remoto**, el §6
del roadmap (ritual por hito) y ADR 0007.

- **Rama**: `h4-applet-boe-puerto`
- **Commit publicado**: `d45f0ea6c99da4a5cc3476a98242e5babbefe7ad` (`feat(H4): T038`), 44 commits por delante de
  `main` (`2c889c4`, igual en local y en `origin` tras `git fetch`, y ancestro de `HEAD`)
- **Propuesta de cambio**: [#24](https://github.com/jmorenobl/kitlegal/pull/24) → `main`, abierta, no borrador,
  `MERGEABLE`; abierta en el intento 1, reutilizada en el 2
- **Fecha**: 2026-09-13 (el run, entre las 21:02 y las 21:06 UTC; 23:02 a 23:06, hora de Madrid)
- **Intentos**: 1 de 3 en rojo (el test de superficie del binario medía solo la plataforma local; causa y arreglo en
  [`tarea-T037.md`](./tarea-T037.md), arreglo en T038); **2 de 3 en verde: `ci` y los cuatro estados de Codecov, todos
  con medida real**

**Lo que esta tarea no hace** (ADR 0007): no fusiona, no empuja a `main`, no fuerza, no borra ramas y no crea etiquetas
ni releases. La fusión es siempre humana.

## Resumen

| Comprobación | Resultado |
|---|---|
| Rama publicada, con el gancho `pre-push` ejecutado | ✅ avance rápido `40efb06..d45f0ea`, `guardia-push` en verde, sin `--force` |
| Propuesta hacia `main` con [`pr-h4.md`](./pr-h4.md) como cuerpo | ✅ #24, ya existía: no se creó otra; cuerpo actualizado al final de este intento |
| Trabajo `ci` (run `34782609287`) | ✅ `success` en 3 m 11 s: `0 issues.`, los dos perfiles de test en `ok`, `internal` incluido, `ci: todos los controles en verde` (§3) |
| `codecov/project` (≥ 70 %) | ✅ **94,73 %** (§4) |
| `codecov/project/internal/core` (≥ 85 %) | ✅ **90,36 %** (§4) |
| `codecov/project/internal/cli` (≥ 90 %) | ✅ **98,09 %** (§4) |
| `codecov/patch` (objetivo `auto` = cobertura de la base) | ✅ **97,49 %** del diff, objetivo 92,95 % (§4) |
| Ningún estado ausente ni en verde sobre cero ficheros | ✅ los cuatro comparan `2c889c4` con `d45f0ea` sobre 62 ficheros y 2 831 líneas (§4) |
| S9 (`boe-cache-rapida` < 200 ms en el ejecutor) | Segunda medida a favor: `internal/app` en `ok` bajo `-race` en los dos perfiles (§5) |
| `make ci` local sobre el árbol final | ✅ `ci: todos los controles en verde` (§6) |
| Sin fusión, sin push a `main`, sin `--force`, sin etiquetas | ✅ |

---

## Aviso de método

El envoltorio de terminal de esta máquina reescribe la salida de `go test`, `git status --porcelain`, `git diff` y
`gh pr checks`. Todo lo que sigue se leyó con el paso directo (`rtk proxy`) y se copia literal, con una excepción
señalada: `git push`, que se ejecutó tal como lo escribe la tarea. Las órdenes que encadenan un `echo` del código de
salida van dentro de `sh -c`, porque la sesión desatendida no aprueba `; echo` suelto.

---

## 1. Publicación de la rama (intento 2)

Antes de empujar: `origin` tenía la rama en `40efb06` (lo que dejó el intento 1), `HEAD` estaba un commit por delante
(`d45f0ea`, T038), y `origin/main` no se había movido.

```
$ rtk proxy git fetch origin
$ rtk proxy git rev-parse main origin/main
2c889c4d7659417f579339e498941da9887b69c3
2c889c4d7659417f579339e498941da9887b69c3
$ rtk proxy git merge-base --is-ancestor origin/main HEAD && echo "origin/main es ancestro de HEAD"
origin/main es ancestro de HEAD
$ rtk proxy git rev-list --count origin/main..HEAD
44
$ rtk proxy git log --oneline origin/h4-applet-boe-puerto..HEAD
d45f0ea feat(H4): T038
```

```
$ git push -u origin h4-applet-boe-puerto
╭─────────────────────────────────────╮
│ 🥊 lefthook v1.13.6  hook: pre-push │
╰─────────────────────────────────────╯
┃  guardia-push ❯

  ────────────────────────────────────
summary: (done in 0.01 seconds)
✔️ guardia-push (0.01 seconds)
To github.com:jmorenobl/kitlegal.git
   40efb06..d45f0ea  h4-applet-boe-puerto -> h4-applet-boe-puerto
branch 'h4-applet-boe-puerto' set up to track 'origin/h4-applet-boe-puerto'.
```

```
$ rtk proxy git rev-parse HEAD origin/h4-applet-boe-puerto
d45f0ea6c99da4a5cc3476a98242e5babbefe7ad
d45f0ea6c99da4a5cc3476a98242e5babbefe7ad
```

Avance rápido (`40efb06..d45f0ea`, sin `+` ni `forced update`). El gancho `pre-push` dejó pasar la referencia: es una
rama de hito, no `main`, no va forzada, no es un borrado ni una etiqueta.

## 2. La propuesta de cambio ya existía

```
$ rtk proxy gh pr view h4-applet-boe-puerto --json number,state,isDraft,baseRefName,headRefName,headRefOid,mergeable,url
{"baseRefName":"main","headRefName":"h4-applet-boe-puerto","headRefOid":"40efb064a5a3f3e6b1ca2e3eae8d751f85d5d878","isDraft":false,"mergeable":"MERGEABLE","number":24,"state":"OPEN","url":"https://github.com/jmorenobl/kitlegal/pull/24"}
```

Como la tarea manda, al encontrar la #24 no se ejecutó `gh pr create`. Tras el push, la propuesta recogió el nuevo
commit sola:

```
$ rtk proxy gh pr view h4-applet-boe-puerto --json number,state,headRefOid,mergeable,url
{"headRefOid":"d45f0ea6c99da4a5cc3476a98242e5babbefe7ad","mergeable":"MERGEABLE","number":24,"state":"OPEN","url":"https://github.com/jmorenobl/kitlegal/pull/24"}
```

El cuerpo es [`pr-h4.md`](./pr-h4.md). Al final de este intento se completó con el resultado de la plataforma (§7) y la
propuesta se actualizó con el mismo fichero (`gh pr edit 24 --body-file …`).

## 3. Integración continua

`gh pr checks --watch` refrescó cada 10 s durante unos tres minutos y terminó con código 0 cuando `ci` pasó. Se
omiten las repeticiones de la línea `pending`:

```
$ rtk proxy sh -c 'gh pr checks h4-applet-boe-puerto --watch 2>&1; echo "código de salida: $?"'
Refreshing checks status every 10 seconds. Press Ctrl+C to quit.

ci	pending	0	https://github.com/jmorenobl/kitlegal/actions/runs/34782609287/job/103792240326
[… 17 refrescos más en pending …]
ci	pass	3m11s	https://github.com/jmorenobl/kitlegal/actions/runs/34782609287/job/103792240326
código de salida: 0
```

```
$ rtk proxy gh run list --branch h4-applet-boe-puerto
completed	success	feat(H4): applet boe, puerto de boe.py	ci	h4-applet-boe-puerto	pull_request	34782609287	3m15s	2026-09-13T21:02:45Z
completed	failure	feat(H4): applet boe, puerto de boe.py	ci	h4-applet-boe-puerto	pull_request	34781187264	2m17s	2026-09-13T20:34:23Z
```

Dos runs: el del intento 1 (rojo, §8) y el de este intento. Solo el flujo `ci` se dispara con una propuesta de cambio:
CodeQL y `nightly` son programados.

```
$ rtk proxy gh run view 34782609287 --json jobs --jq '.jobs[] | "JOB \(.name)\t\(.conclusion)\t\(.startedAt)\t\(.completedAt)", (.steps[] | "  \(.number)\t\(.name)\t\(.conclusion)\t\(.startedAt)\t\(.completedAt)")'
JOB ci	success	2026-09-13T21:02:48Z	2026-09-13T21:05:59Z
  1	Set up job	success	2026-09-13T21:02:49Z	2026-09-13T21:02:50Z
  2	Obtener el código	success	2026-09-13T21:02:50Z	2026-09-13T21:02:51Z
  3	Instalar Go y restaurar la caché	success	2026-09-13T21:02:51Z	2026-09-13T21:03:17Z
  4	Ejecutar los controles	success	2026-09-13T21:03:17Z	2026-09-13T21:05:54Z
  5	Publicar el perfil de cobertura	success	2026-09-13T21:05:54Z	2026-09-13T21:05:57Z
  9	Post Instalar Go y restaurar la caché	success	2026-09-13T21:05:57Z	2026-09-13T21:05:57Z
  10	Post Obtener el código	success	2026-09-13T21:05:57Z	2026-09-13T21:05:57Z
  11	Complete job	success	2026-09-13T21:05:57Z	2026-09-13T21:05:57Z
```

En el intento 1, el paso 5 salió `skipped` tras el fallo del 4. Aquí los ocho pasos están en `success`. La caché de
`setup-go` se restauró con la misma clave que entonces (H4 no cambia `go.sum` frente a `main`):

```
Instalar Go y restaurar la caché	2026-09-13T21:03:17.2938143Z Cache restored successfully
Instalar Go y restaurar la caché	2026-09-13T21:03:17.3283756Z Cache restored from key: setup-go-Linux-x64-ubuntu24-go-1.27.1-faad8ae2390159a627abd187a613292da7f4afabaf9cffc8224f62e16a4222ae
```

Registro del paso «Ejecutar los controles» (`gh run view 34782609287 --log`, 611 líneas, filtrado por los resultados de
cada control; se quita el prefijo del trabajo):

```
Ejecutar los controles	2026-09-13T21:03:18.1666184Z go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
Ejecutar los controles	2026-09-13T21:03:20.3260106Z go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
Ejecutar los controles	2026-09-13T21:03:51.4866981Z 0 issues.
Ejecutar los controles	2026-09-13T21:03:51.5571601Z go test -race -shuffle=on -coverprofile=coverage.out ./...
Ejecutar los controles	2026-09-13T21:03:56.6288201Z ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	1.020s	coverage: 0.0% of statements
Ejecutar los controles	2026-09-13T21:03:58.3566113Z ok  	github.com/jmorenobl/kitlegal/internal	2.230s	coverage: [no statements]
Ejecutar los controles	2026-09-13T21:04:03.7560543Z ok  	github.com/jmorenobl/kitlegal/internal/app	5.623s	coverage: 93.7% of statements
Ejecutar los controles	2026-09-13T21:04:07.2926736Z ok  	github.com/jmorenobl/kitlegal/internal/cache	5.751s	coverage: 90.7% of statements
Ejecutar los controles	2026-09-13T21:04:07.2976530Z ok  	github.com/jmorenobl/kitlegal/internal/cli	1.606s	coverage: 98.6% of statements
Ejecutar los controles	2026-09-13T21:04:07.3008082Z ?   	github.com/jmorenobl/kitlegal/internal/core	[no test files]
Ejecutar los controles	2026-09-13T21:04:09.1987905Z ok  	github.com/jmorenobl/kitlegal/internal/core/schema	1.037s	coverage: 90.1% of statements
Ejecutar los controles	2026-09-13T21:04:42.6180098Z ok  	github.com/jmorenobl/kitlegal/internal/httpx	31.386s	coverage: 97.2% of statements
Ejecutar los controles	2026-09-13T21:04:42.6183105Z ok  	github.com/jmorenobl/kitlegal/internal/render	1.016s	coverage: 95.8% of statements
Ejecutar los controles	2026-09-13T21:04:42.6184267Z ok  	github.com/jmorenobl/kitlegal/internal/source/boe	4.401s	coverage: 99.4% of statements
Ejecutar los controles	2026-09-13T21:04:42.6477371Z go test -race -tags=integration -coverprofile=coverage-integration.out ./...
Ejecutar los controles	2026-09-13T21:04:44.9675949Z ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	1.016s	coverage: 0.0% of statements
Ejecutar los controles	2026-09-13T21:04:46.5098742Z ok  	github.com/jmorenobl/kitlegal/internal	2.057s	coverage: [no statements]
Ejecutar los controles	2026-09-13T21:04:50.3139676Z ok  	github.com/jmorenobl/kitlegal/internal/app	5.308s	coverage: 93.7% of statements
Ejecutar los controles	2026-09-13T21:04:56.1306526Z ok  	github.com/jmorenobl/kitlegal/internal/cache	6.932s	coverage: 96.2% of statements
Ejecutar los controles	2026-09-13T21:04:56.1310415Z ok  	github.com/jmorenobl/kitlegal/internal/cli	1.631s	coverage: 98.6% of statements
Ejecutar los controles	2026-09-13T21:04:56.1311830Z ?   	github.com/jmorenobl/kitlegal/internal/core	[no test files]
Ejecutar los controles	2026-09-13T21:04:56.1329153Z ok  	github.com/jmorenobl/kitlegal/internal/core/schema	1.039s	coverage: 90.1% of statements
Ejecutar los controles	2026-09-13T21:05:25.9878350Z ok  	github.com/jmorenobl/kitlegal/internal/httpx	30.320s	coverage: 97.2% of statements
Ejecutar los controles	2026-09-13T21:05:25.9885978Z ok  	github.com/jmorenobl/kitlegal/internal/render	(cached)	coverage: 95.8% of statements
Ejecutar los controles	2026-09-13T21:05:25.9889176Z ok  	github.com/jmorenobl/kitlegal/internal/source/boe	4.406s	coverage: 99.4% of statements
Ejecutar los controles	2026-09-13T21:05:31.7785634Z No vulnerabilities found.
Ejecutar los controles	2026-09-13T21:05:31.7892212Z go test -count=1 -run '^TestEsquemasPublicados$' ./internal/app/
Ejecutar los controles	2026-09-13T21:05:42.1275017Z ok  	github.com/jmorenobl/kitlegal/internal/app	0.846s
Ejecutar los controles	2026-09-13T21:05:49.2197032Z 9:05PM INF no leaks found
Ejecutar los controles	2026-09-13T21:05:50.9675707Z all modules verified
Ejecutar los controles	2026-09-13T21:05:52.6207868Z all modules verified
Ejecutar los controles	2026-09-13T21:05:54.3672987Z all modules verified
Ejecutar los controles	2026-09-13T21:05:54.5679547Z all modules verified
Ejecutar los controles	2026-09-13T21:05:54.6238952Z all modules verified
Ejecutar los controles	2026-09-13T21:05:54.6247868Z go mod tidy -diff
Ejecutar los controles	2026-09-13T21:05:54.6892476Z ci: todos los controles en verde
```

(Se han quitado los códigos de color de la línea `INF` de `gitleaks`.) Ninguna línea `FAIL`, `DATA RACE` ni
`make: ***` en el registro. Lo que cambia respecto al intento 1 es exactamente lo que T038 arregló: `internal` pasa de
`FAIL … 0.285s` a `ok … 2.230s` en el ejecutor linux, porque `TestDependenciasDelBinario` mide ahora las seis
plataformas de distribución (seis `go list -deps` con `GOOS`, `GOARCH` y `CGO_ENABLED=0`) y compara la unión con la
lista declarada, que no cambió. Y por primera vez en esta propuesta corren en remoto `test-integration`, `vuln`,
`schema-check`, `secrets`, `mod-verify` (la raíz y los cuatro módulos de herramienta) y `mod-tidy-check`, todos en
verde.

## 4. Codecov: los cuatro estados, emitidos y con medida real

Subida del perfil (paso 5, `codecov/codecov-action`, filtrado):

```
  files: ./coverage.out,./coverage-integration.out
  fail_ci_if_error: true
  handle_no_reports_found: false
==> Token set from input
      ./codecov  upload-coverage -t <redacted> --fail-on-error --git-service github --sha d45f0ea6c99da4a5cc3476a98242e5babbefe7ad --file ./coverage.out --file ./coverage-integration.out --gcov-executable gcov
info - 2026-09-13 21:05:56,494 -- ci service found: github-actions
info - 2026-09-13 21:05:56,614 -- Found 2 coverage files to report
info - 2026-09-13 21:05:56,615 -- > /home/runner/work/kitlegal/kitlegal/coverage.out
info - 2026-09-13 21:05:56,615 -- > /home/runner/work/kitlegal/kitlegal/coverage-integration.out
info - 2026-09-13 21:05:57,053 -- Your upload is now queued for processing. When finished, results will be available at: https://app.codecov.io/github/jmorenobl/kitlegal/commit/d45f0ea6c99da4a5cc3476a98242e5babbefe7ad
info - 2026-09-13 21:05:57,054 -- Sending upload (210300 bytes) to storage
info - 2026-09-13 21:05:57,187 -- Upload queued for processing complete
```

Los dos perfiles que `ci.yml` sube (el unitario y el de `-tags=integration`) llegaron a Codecov, que los une. Los
estados aparecieron entre 7 y 10 s después de que `ci` terminara, o sea, **después** de que `gh pr checks --watch`
saliera con 0; por eso se leyeron aparte, por el título de cada check-run:

```
$ rtk proxy gh api repos/jmorenobl/kitlegal/commits/d45f0ea6c99da4a5cc3476a98242e5babbefe7ad/check-runs --jq '.check_runs[] | "\(.name)\t\(.status)\t\(.conclusion)\t\(.started_at)\t\(.completed_at)\t\(.output.title)"'
codecov/project/internal/cli	completed	success	2026-09-13T21:06:09Z	2026-09-13T21:06:09Z	98.09% (target 90.00%)
codecov/project/internal/core	completed	success	2026-09-13T21:06:08Z	2026-09-13T21:06:08Z	90.36% (target 85.00%)
codecov/patch	completed	success	2026-09-13T21:06:07Z	2026-09-13T21:06:08Z	97.49% of diff hit (target 92.95%)
codecov/project	completed	success	2026-09-13T21:06:06Z	2026-09-13T21:06:07Z	94.73% (target 70.00%)
ci	completed	success	2026-09-13T21:02:48Z	2026-09-13T21:05:59Z	null
```

```
$ rtk proxy gh pr checks h4-applet-boe-puerto
ci	pass	3m11s	https://github.com/jmorenobl/kitlegal/actions/runs/34782609287/job/103792240326
codecov/patch	pass	1s	https://app.codecov.io/gh/jmorenobl/kitlegal/pull/24
codecov/project	pass	1s	https://app.codecov.io/gh/jmorenobl/kitlegal/pull/24
codecov/project/internal/cli	pass	0	https://app.codecov.io/gh/jmorenobl/kitlegal/pull/24
codecov/project/internal/core	pass	0	https://app.codecov.io/gh/jmorenobl/kitlegal/pull/24
```

`statusCheckRollup` de la #24 (`gh pr view 24 --json statusCheckRollup`): cinco check-runs, los cinco `COMPLETED` y
`SUCCESS`; `mergeable: MERGEABLE`.

| Estado | Umbral (`codecov.yml`) | Medido | Verde vacío | Referencia local (T036, `pr-h4.md`) |
|---|---|---|---|---|
| `codecov/project` | ≥ 70 % | **94,73 %** | No: 2 682 de 2 831 líneas | 96,1 % (3 109/3 236 sentencias, unión de los dos perfiles) |
| `codecov/project/internal/core` | ≥ 85 % | **90,36 %** | No: `main` ya contiene `internal/core/**`; la base `2c889c4` se filtra por sus rutas y compara | 90,1 % (73/81 sentencias) |
| `codecov/project/internal/cli` | ≥ 90 % | **98,09 %** | No: `main` ya contiene `internal/cli/**` | 98,6 % (348/353 sentencias) |
| `codecov/patch` | `auto` = cobertura de la base | **97,49 %** del diff, objetivo **92,95 %** (la cobertura de `main` en `2c889c4`) | No: 22 líneas del diff sin cubrir, nombradas | — (T036 lo dejó para leerlo aquí) |

- **Ninguno es un verde sobre cero ficheros.** Los cuatro traen porcentaje y objetivo en el título, y el informe de
  Codecov compara la base `2c889c4` con el head `d45f0ea` sobre 62 ficheros (43 en `main`, 19 nuevos) y 2 831 líneas
  (929 más). Los componentes `internal_core` e `internal_cli` miden de verdad, como anticipó `codecov.yml`: H4 no crea
  ninguno de los dos árboles.
- **Las cifras locales y las de Codecov no son la misma unidad.** `go tool cover -func` cuenta sentencias del perfil;
  Codecov las traduce a líneas de fichero. Por eso el global local es 96,1 % y el remoto 94,73 %, y `internal/cli` 98,6 %
  frente a 98,09 %. Cada umbral queda holgado en las dos medidas, y `codecov.yml` no está en el diff frente a `main`.
- **`codecov/patch` con objetivo `auto`** midió lo que el roadmap pide (no retroceder): el diff de H4 (97,49 %) cubre
  más que la base entera (92,95 %). El comentario de Codecov en la #24 marca la línea del parche con una cruz porque
  hay 22 líneas del diff sin cubrir; el estado que bloquea es el check-run, y está en `success`.

Comentario de `codecov[bot]` en la #24 (21:06:10Z), la parte con datos:

```
Patch coverage is 97.49431% with 22 lines in your changes missing coverage. Please review.
Project coverage is 94.73%. Comparing base (2c889c4) to head (d45f0ea).

| Files with missing lines                       | Patch % | Lines      |
| internal/app/ejemplo/kitlegal-e2e/main.go      |   0.00% | 13 Missing |
| internal/source/boe/fuente.go                  |  98.14% |  3 Missing |
| cmd/kitlegal/main.go                           |   0.00% |  2 Missing |
| internal/source/boe/articulo.go                |  97.80% |  2 Missing |
| internal/app/boe.go                            |  98.00% |  1 Missing |
| internal/app/registro.go                       |  75.00% |  1 Missing |

@@            Coverage Diff             @@
##             main      #24      +/-   ##
==========================================
+ Coverage   92.95%   94.73%   +1.78%
==========================================
  Files          43       62      +19
  Lines        1902     2831     +929
==========================================
+ Hits         1768     2682     +914
- Misses        134      149      +15
```

Las 22 líneas sin cubrir son las dos `main` (13 en el binario de e2e y 2 en el distribuido, que se ejecutan como
procesos y no bajo el perfil) y siete repartidas entre `fuente.go` (3), `articulo.go` (2), `boe.go` (1) y
`registro.go` (1). No se añade ninguna exclusión para esconderlas.

## 5. S9: segunda medida de `boe-cache-rapida` en el ejecutor

S9 (research D20) supone que las diez invocaciones con caché caliente bajan de 200 ms en el ejecutor de la plataforma,
con la suite en paralelo y `-race`. En este run, `internal/app` quedó en `ok` en los dos perfiles: 5,623 s con
`-race -shuffle=on` y 5,308 s con `-race -tags=integration`. Ese paquete ejecuta `TestEntregaDelHito`, cuyos guiones
incluyen `boe-cache-rapida`, sin restricción de compilación, sin `t.Skip` y sin condición de salto en el `.txtar`
(comprobado en el intento 1, §8). Con las del intento 1, son **tres ejecuciones del paquete en el ejecutor y las tres
en verde**. La reserva se mantiene tal cual: sin `-v`, el registro no nombra el subtest; ninguna ejecución con `-v` en
el ejecutor forma parte de este hito.

## 6. `make ci` final de este intento

Se ejecutó en primer plano sobre el árbol final: `d45f0ea` más los cambios del directorio del feature de este intento
(esta evidencia, `pr-h4.md` y la nota de `tarea-T037.md`), ya escritos, así que `gitleaks` los recorrió. T037 no toca
código.

```
$ rtk proxy sh -c 'make ci > /tmp/kitlegal-t037-make-ci-2.log 2>&1; echo "código $?"; grep -E "^(ok|FAIL|---|ci:|\?)|issues|DATA RACE|vulnerabilities|leaks|all modules|tidy" /tmp/kitlegal-t037-make-ci-2.log'
```

```
código 0
0 issues.
ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	2.897s	coverage: 0.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal	2.355s	coverage: [no statements]
ok  	github.com/jmorenobl/kitlegal/internal/app	4.838s	coverage: 93.7% of statements
ok  	github.com/jmorenobl/kitlegal/internal/cache	4.110s	coverage: 90.7% of statements
ok  	github.com/jmorenobl/kitlegal/internal/cli	2.578s	coverage: 98.6% of statements
?   	github.com/jmorenobl/kitlegal/internal/core	[no test files]
ok  	github.com/jmorenobl/kitlegal/internal/core/schema	3.002s	coverage: 90.1% of statements
ok  	github.com/jmorenobl/kitlegal/internal/httpx	16.189s	coverage: 97.2% of statements
ok  	github.com/jmorenobl/kitlegal/internal/render	3.634s	coverage: 95.8% of statements
ok  	github.com/jmorenobl/kitlegal/internal/source/boe	2.941s	coverage: 99.4% of statements
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
ok  	github.com/jmorenobl/kitlegal/internal/app	0.865s
go tool -modfile=tools/gitleaks/go.mod gitleaks dir . --redact --no-banner
11:14PM INF no leaks found
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

- `internal` en `ok` (2,355 s) también en local: con T038, el test mide las seis plataformas desde cualquier máquina, y
  el veredicto local y el remoto coinciden por primera vez en esta propuesta.
- **`test-integration` sale `(cached)`**, y es legítimo: desde `d45f0ea` no ha cambiado ningún fichero Go ni material de
  test; solo el directorio del feature.
- **Veredicto del intento 2**: la verificación determinista local está en verde y el criterio de la tarea (el estado
  de la integración continua y de Codecov en remoto) se cumple con medida real en los cinco estados. **T037 se marca
  `[X]`.**

## 7. Cambios de este intento, todo dentro del directorio del feature

- Esta evidencia, reescrita con el intento 2 delante y el intento 1 como histórico (§8).
- `pr-h4.md`, el cuerpo de la #24: «Alcance» con las cifras sobre `d45f0ea` y el cambio de T038 en
  `internal/arch_test.go`; «Controles añadidos» con la medida en las seis plataformas; «Evidencia» con un apartado de
  plataforma (los cinco estados con sus cifras); «Pendientes» con S9 actualizado, el rojo del intento 1 cerrado por T038
  y T037 con su resultado. La propuesta se actualizó con el fichero (`gh pr edit 24 --body-file …`).
- `tarea-T037.md`: una nota al principio con el resultado del intento 2; el análisis del intento 1 no cambia.
- `tasks.md`: T037 marcada `[X]`, solo tras el `make ci` de §6 en verde.

## 8. Intento 1 (histórico): `ci` en rojo por un test que solo medía la plataforma local

Base `40efb06` (`feat(H4): T036`). El análisis completo, el diseño del arreglo y su comprobación sobre un clon están
en [`tarea-T037.md`](./tarea-T037.md); aquí queda lo que la plataforma dijo.

- **Publicación**: `git ls-remote --heads origin h4-applet-boe-puerto` no devolvía nada; `git push -u origin
  h4-applet-boe-puerto` con `guardia-push` en verde y `* [new branch] h4-applet-boe-puerto -> h4-applet-boe-puerto`.
- **Propuesta**: `gh pr view h4-applet-boe-puerto` → `no pull requests found for branch "h4-applet-boe-puerto"`;
  `gh pr create --base main --head h4-applet-boe-puerto --title "feat(H4): applet boe, puerto de boe.py" --body-file
  specs/005-h4-applet-boe-puerto/gates/pr-h4.md` → `https://github.com/jmorenobl/kitlegal/pull/24`.
- **`ci`** (run `34781187264`, `failure`, 2 m 13 s): `formato` y `lint` en verde (`0 issues.`), todos los paquetes en `ok`
  salvo `internal`:

```
ci	Ejecutar los controles	2026-09-13T20:35:48.1558467Z --- FAIL: TestDependenciasDelBinario (0.13s)
ci	Ejecutar los controles	2026-09-13T20:35:48.1651662Z         	            	extra elements in list A:
ci	Ejecutar los controles	2026-09-13T20:35:48.1665531Z         	            	 (string) (len=26) "github.com/mattn/go-isatty",
ci	Ejecutar los controles	2026-09-13T20:35:48.1681948Z         	            	 (string) (len=30) "github.com/ncruces/go-strftime"
ci	Ejecutar los controles	2026-09-13T20:35:48.1976576Z FAIL	github.com/jmorenobl/kitlegal/internal	0.285s
ci	Ejecutar los controles	2026-09-13T20:36:36.1249216Z make: *** [Makefile:67: test] Error 1
```

- **Codecov no emitió ningún estado, y la causa no era la configuración**: el paso «Publicar el perfil de cobertura»
  salió `skipped` porque GitHub Actions no ejecuta los pasos siguientes a uno que ha fallado. `check-runs` del commit
  `40efb06` solo tenía `ci … failure`. Con `make ci` en verde (§3 y §4), los cuatro estados llegaron sin tocar
  `codecov.yml` ni `ci.yml`, que es lo que el intento 1 anticipó.
- **Causa**: `TestDependenciasDelBinario` (T026) lanzaba `go list -deps` con el entorno del proceso, así que medía solo
  la plataforma del ordenador que ejecuta el test, y los módulos que enlaza el binario dependen de ella
  (`modernc.org/libc` elige sus ficheros por sistema): 18 en darwin, 16 en linux (sin `go-isatty` ni `go-strftime`),
  17 en windows (sin `uuid`). La lista declarada, la unión de las seis plataformas de distribución, era correcta. Se
  reprodujo en local con `go test -exec 'env GOOS=linux GOARCH=amd64 CGO_ENABLED=0' -run '^TestDependenciasDelBinario$'
  ./internal/` sobre un clon de `40efb06`.
- **Por qué no se arregló en T037**: `internal/arch_test.go` está fuera de sus rutas, y redelimitarla no servía porque
  el workflow commitea después del intento y el push ocurre durante él. El arreglo fue **T038**, sin etiqueta de
  plataforma y colocada antes de T037 en `tasks.md`: mide las seis plataformas con `GOOS`, `GOARCH` y `CGO_ENABLED=0`,
  compara la unión con `modulosDelBinario` y falla tanto por un módulo sin declarar en cualquier plataforma como por uno
  declarado que no enlace en ninguna. Comprobado sobre el clon con las sondas negativas y `make ci` en verde, y
  commiteado en `d45f0ea`.
- **S9, primera medida**: `internal/app` en `ok` a los 5,098 s bajo `-race -shuffle=on`; `e2e_test.go` sin restricción
  de compilación ni `t.Skip` (`grep -c t.Skip` → `0`), `boe-cache-rapida.txtar` sin condición de salto.
- **`make ci` local del intento 1**: código 0, `ci: todos los controles en verde` en darwin/arm64, donde el binario
  enlaza los 18 módulos. Ese verde local no sustituía al remoto, y por eso T037 quedó `[ ]`.
