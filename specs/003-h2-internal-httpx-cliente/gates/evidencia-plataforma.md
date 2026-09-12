# Evidencia de plataforma de H2 (T019)

Publicación de la rama del hito, apertura de la propuesta de cambio y lectura de los estados de la
integración continua y de Codecov. Cubre el punto 1 de la Definition of Done leído **en remoto**, §6 del
roadmap (ritual por hito, paso 3) y ADR 0007.

- **Rama**: `h2-internal-httpx-cliente`
- **Commit publicado**: `f18cfc5a6f9170c3fc4eb2dbd94c3bf76622ffec` (`feat(H2): T018`)
- **Propuesta de cambio**: [#15](https://github.com/jmorenobl/kitlegal/pull/15) → `main`
- **Fecha**: 2026-09-12
- **Intento**: 1 de 3

**Lo que esta tarea no hace** (ADR 0007, constitución §4 del flujo de trabajo): no fusiona, no empuja a
`main`, no fuerza, no borra ramas y no crea etiquetas ni releases. La fusión (squash-merge) es siempre
humana.

---

## Aviso de método: el envoltorio del terminal también reescribe `gh pr checks`

La evidencia de cierre ya documenta que el envoltorio de esta máquina reescribe `go test`, `git status
--porcelain` y `git diff`. Al leer los estados remotos se comprobó que **también reescribe `gh pr
checks`**, y con un resumen que no se corresponde con la realidad:

```
$ gh pr checks h2-internal-httpx-cliente --watch --interval 20
CI Checks Summary:
  [ok] Passed: 2
  [FAIL] Failed: 0
  [pending] Pending: 14
```

La propuesta tiene **cinco** estados en total, no dieciséis, y en ese momento los cinco estaban ya en
verde. Ese resumen no es una lectura de la plataforma: ni los recuentos ni la suma cuadran. Una tarea que
se fiara de él leería «14 pendientes» sobre una propuesta terminada —o, en el caso simétrico, «0 fallos»
sobre una en rojo—.

**Todo lo que sigue está leído con el paso directo (`rtk proxy`).** Es la misma cautela que el escenario
12 del quickstart impone a las medidas locales, extendida a la lectura del estado remoto.

---

## 1. Publicación de la rama

```
$ git push -u origin h2-internal-httpx-cliente
╭─────────────────────────────────────╮
│ 🥊 lefthook v1.13.6  hook: pre-push │
╰─────────────────────────────────────╯
┃  guardia-push ❯
✔️ guardia-push (0.01 seconds)
remote:
remote: Create a pull request for 'h2-internal-httpx-cliente' on GitHub by visiting:
remote:      https://github.com/jmorenobl/kitlegal/pull/new/h2-internal-httpx-cliente
remote:
To github.com:jmorenobl/kitlegal.git
 * [new branch]      h2-internal-httpx-cliente -> h2-internal-httpx-cliente
branch 'h2-internal-httpx-cliente' set up to track 'origin/h2-internal-httpx-cliente'.
```

El gancho `pre-push` (`scripts/lefthook/pre-push/guardia-push.sh`) se ejecutó y dejó pasar la referencia:
es una rama de hito, no `main`, no forzada, no un borrado y no una etiqueta. Rama nueva en `origin`: 23
commits por delante de `main`, ninguno de ellos reescrito.

## 2. Apertura de la propuesta de cambio

No existía propuesta previa para la rama:

```
$ gh pr view h2-internal-httpx-cliente
no pull requests found for branch "h2-internal-httpx-cliente"
```

Creada con la plantilla del ritual (§6.3 del roadmap: objetivo, alcance, controles añadidos, decisiones y
pendientes), más la justificación de dependencias que exige la constitución §V y el bloque de evidencia:

```
$ gh pr create --base main --head h2-internal-httpx-cliente \
    --title 'feat(H2): internal/httpx, cliente HTTP responsable y grabación de fixtures' \
    --body-file specs/003-h2-internal-httpx-cliente/gates/pr-cuerpo.md
https://github.com/jmorenobl/kitlegal/pull/15
```

El cuerpo vive en [`pr-cuerpo.md`](./pr-cuerpo.md), en el directorio del feature, y adjunta como evidencia
[`evidencia-sc012.md`](./evidencia-sc012.md) (los controles fallan ante cada intento de saltárselos) y
[`evidencia-cierre.md`](./evidencia-cierre.md) (validación agregada del hito).

**Dependencias (constitución §V), tal como van en el cuerpo**: las dos entradas nuevas de `go.mod`
—`github.com/temoto/robotstxt v1.1.2` y `golang.org/x/time v0.16.0`— están **en la lista fijada** de la
constitución §V y de `docs/ROADMAP.md` §3, así que ninguna requiere *Complexity Tracking*. No se añade
ninguna fuera de la lista.

## 3. Integración continua

```
$ gh run list --branch h2-internal-httpx-cliente
completed	success	feat(H2): internal/httpx, cliente HTTP responsable y grabación de fixtures	ci	h2-internal-httpx-cliente	pull_request	34684448466	4m42s	2026-09-12T08:54:05Z

$ gh run view 34684448466
✓ h2-internal-httpx-cliente ci jmorenobl/kitlegal#15 · 34684448466
Triggered via pull_request about 5 minutes ago

JOBS
✓ ci in 4m39s (ID 103528870146)
```

Un único flujo (`ci`), un único trabajo, en verde en 4 m 39 s. Es el mismo `make ci` que la batería por
tarea ejecuta en local: formato, lint, tests con `-race`, vulnerabilidades, esquemas, secretos, integridad
de módulos y `go mod tidy -diff`. El punto 1 de la Definition of Done queda acreditado también en remoto,
sobre una máquina limpia y sin red del desarrollo.

## 4. Estados de la propuesta

```
$ rtk proxy gh pr checks h2-internal-httpx-cliente
ci	pass	4m39s	https://github.com/jmorenobl/kitlegal/actions/runs/34684448466/job/103528870146
codecov/patch	pass	0	https://app.codecov.io/gh/jmorenobl/kitlegal/pull/15
codecov/project	pass	0	https://app.codecov.io/gh/jmorenobl/kitlegal/pull/15
codecov/project/internal/cli	pass	0	https://app.codecov.io/gh/jmorenobl/kitlegal/pull/15
codecov/project/internal/core	pass	0	https://app.codecov.io/gh/jmorenobl/kitlegal/pull/15
```

Cinco estados, cinco en verde, ninguno pendiente y ninguno ausente.

## 5. Que ningún estado de Codecov es un «verde vacío» (S3, `codecov.yml`)

`codecov.yml` documenta el modo de fallo que había que descartar: Codecov compara el head con el informe
de la **base** filtrado por los paths del componente y, si en la base no hay ningún fichero de ese árbol,
el estado sale en verde **sin evaluar el objetivo** («No coverage information found on base report»). Eso
es lo que le pasó a `internal/core` en H0 y a `internal/cli` en la propuesta de H1, que **crea** el árbol.

La lectura del resumen (`pass`) no distingue un caso del otro. Lo que los distingue es el título del
estado, que solo trae porcentaje y objetivo cuando ha medido de verdad:

```
$ rtk proxy gh api repos/jmorenobl/kitlegal/commits/f18cfc5a6f9170c3fc4eb2dbd94c3bf76622ffec/check-runs \
    --jq '.check_runs[] | "\(.name)\t\(.conclusion)\t\(.output.title)"'
codecov/project/internal/cli	success	98.07% (target 90.00%)
codecov/project/internal/core	success	90.36% (target 85.00%)
codecov/patch	success	94.83% of diff hit (target 89.52%)
codecov/project	success	91.91% (target 70.00%)
ci	success	null
```

| Estado | Medido | Objetivo | ¿Midió de verdad? |
|---|---|---|---|
| `codecov/project` | **91,91 %** | 70 % | ✅ porcentaje y objetivo en el título |
| `codecov/project/internal/core` | **90,36 %** | 85 % | ✅ la base ya contiene `internal/core/**` desde H0/H1 |
| `codecov/project/internal/cli` | **98,07 %** | 90 % | ✅ la base ya contiene `internal/cli/**` desde H1 |
| `codecov/patch` | **94,83 %** del diff | 89,52 % | ✅ mide el diff de esta propuesta, que son 14 733 líneas |

Ninguno dice «No coverage information found on base report». Los dos componentes miden por primera vez en
una propuesta **posterior** a la que creó su árbol, que es justo lo que `codecov.yml` anticipaba y lo que
H1 demostró con sondas. Y `codecov/patch` sobre 14 733 líneas añadidas no puede ser un verde sobre cero
ficheros.

**H2 no declara componente nuevo** (S3 de `research.md`): rigen los umbrales generales, así que no hay
ningún verde vacío que demostrar con sondas en este hito.

### Cotejo con la medida local

| Umbral | Exigido | Local (`evidencia-cierre.md`) | Codecov |
|---|---|---|---|
| Global | ≥ 70 % | 93,3 % | **91,91 %** |
| `internal/core/**` | ≥ 85 % | 90,1 % | **90,36 %** |
| `internal/cli` | ≥ 90 % | 98,6 % | **98,07 %** |

Las cifras no coinciden al decimal y no tienen por qué: la orden es la misma (`go test -race -shuffle=on
-coverprofile=coverage.out ./...`, `Makefile:67`, idéntica en local y en el flujo remoto) sobre el mismo
perfil, pero `go tool cover -func` cuenta **sentencias** y Codecov recuenta el perfil por **líneas**. Lo
que importa para el punto 9 de la Definition of Done es que **las dos lecturas superan los tres umbrales
por separado** y que ningún objetivo se rebajó: `codecov.yml` no aparece en el diff frente a `main` y
conserva sus tres objetivos (`70%`, `85%`, `90%`) con `informational: false`.

---

## Veredicto

| Criterio | Estado |
|---|---|
| Rama del hito publicada en `origin`, con el gancho `pre-push` ejecutado | ✅ |
| Propuesta de cambio abierta hacia `main` con la plantilla del ritual y la justificación §V | ✅ |
| Evidencia de SC-012 y de cierre adjunta en el cuerpo | ✅ |
| Integración continua en verde en remoto (`ci`, run `34684448466`) | ✅ |
| Los cuatro estados de Codecov en verde **y midiendo de verdad** (no verde vacío) | ✅ |
| Ningún umbral rebajado | ✅ |
| Sin fusión, sin push a `main`, sin `--force`, sin etiquetas | ✅ |

**Pendiente humano**: la revisión del ritual (§6.4: `/code-review` y `/security-review` sobre la propuesta)
y el squash-merge. Ninguno de los dos es del workflow.

## `make ci` final

Ejecutado al terminar la tarea, con este fichero y `pr-cuerpo.md` ya escritos. T019 no toca ni una línea de
código: sus únicos ficheros son los dos de `gates/` y la marca en `tasks.md`, así que el resultado tiene que
ser el mismo que el del cierre —y lo es.

```
$ rtk proxy make ci
go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
0 issues.
go test -race -shuffle=on -coverprofile=coverage.out ./...
ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	1.372s	coverage: 0.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal	1.867s	coverage: [no statements]
ok  	github.com/jmorenobl/kitlegal/internal/app	3.499s	coverage: 92.8% of statements
	github.com/jmorenobl/kitlegal/internal/app/ejemplo		coverage: 0.0% of statements
	github.com/jmorenobl/kitlegal/internal/app/ejemplo/kitlegal-e2e		coverage: 0.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal/cli	2.316s	coverage: 98.6% of statements
ok  	github.com/jmorenobl/kitlegal/internal/core/schema	2.524s	coverage: 90.1% of statements
ok  	github.com/jmorenobl/kitlegal/internal/httpx	15.678s	coverage: 94.8% of statements
ok  	github.com/jmorenobl/kitlegal/internal/render	2.955s	coverage: 95.8% of statements
go tool -modfile=tools/govulncheck/go.mod govulncheck ./...
No vulnerabilities found.
schema-check: no hay schemas/ todavía; los aportan H4 (borrador) y H10 (contrato)
go tool -modfile=tools/gitleaks/go.mod gitleaks dir . --redact --no-banner
INF scanned ~7187225 bytes (7.19 MB) in 667ms
INF no leaks found
go mod verify
all modules verified
== tools/gitleaks
all modules verified
== tools/golangci-lint
all modules verified
== tools/govulncheck
all modules verified
== tools/lefthook
all modules verified
go mod tidy -diff
ci: todos los controles en verde
```

Las cifras de cobertura son las mismas que en `evidencia-cierre.md`, como corresponde a una tarea que no
añade producto. **T019 queda marcada.**

**Ficheros que toca T019**: `specs/003-h2-internal-httpx-cliente/gates/pr-cuerpo.md` (el cuerpo de la
propuesta), este fichero y la marca de T019 en `specs/003-h2-internal-httpx-cliente/tasks.md`. Ningún
fichero de producción, nada bajo `testdata/` ni `schemas/`.
