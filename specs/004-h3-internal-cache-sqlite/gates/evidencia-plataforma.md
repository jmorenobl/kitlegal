# Evidencia de plataforma de H3 (T014)

Publicación de la rama del hito, apertura de la propuesta de cambio y lectura de los estados de la
integración continua y de Codecov. Cubre el punto 1 de la Definition of Done leído **en remoto**, el §6
del roadmap (ritual por hito, paso 3), ADR 0007 y la obligación 13 del plan (supuestos S3 y S5).

- **Rama**: `h3-internal-cache-sqlite`
- **Commit publicado**: `0b1bf84e40ed3256291ee12e7fd686569d3fdf76` (`feat(H3): T013`), 15 commits por
  delante de `main` (`ed6d1c4`)
- **Propuesta de cambio**: [#18](https://github.com/jmorenobl/kitlegal/pull/18) → `main`
- **Fecha**: 2026-09-12
- **Intento**: 1 de 3

**Lo que esta tarea no hace** (ADR 0007, constitución «Flujo de trabajo y gates», punto 4): no fusiona, no
empuja a `main`, no fuerza, no borra ramas y no crea etiquetas ni releases. La fusión (squash-merge) es
siempre humana.

## Resumen

| Comprobación | Resultado |
|---|---|
| Rama publicada, con el gancho `pre-push` ejecutado | ✅ `* [new branch]`, `guardia-push` en verde |
| Propuesta abierta con la plantilla del ritual y la justificación §V | ✅ #18, cuerpo en [`pr-h3.md`](./pr-h3.md) |
| Diferencia de indirectos frente a S1, anotada en el cuerpo | ✅ `golang.org/x/tools` v0.26.0 → v0.48.0 |
| **S3**: trabajo `ci` en verde con `test-integration` dentro | ✅ run `34718912959`, `success` |
| **S5**: duración de `ci` de la propuesta | 6 m 19 s, **en frío** (`Cache is not found`); sin medida en caliente en esta tarea (§4) |
| `codecov/project` (global ≥ 70 %) | ✅ **90,99 %**, medido |
| `codecov/project/internal/core` (≥ 85 %) | ✅ **90,36 %**, medido |
| `codecov/project/internal/cli` (≥ 90 %) | ✅ **98,07 %**, medido |
| `codecov/patch` | ❌ **85,64 % del diff** frente a un objetivo **automático** de 92,47 %; no lo declara `codecov.yml` ni lo pide la Definition of Done; queda para decisión humana (§6) |
| Sin fusión, sin push a `main`, sin `--force`, sin etiquetas | ✅ |

---

## Aviso de método

El envoltorio de terminal de esta máquina reescribe `go test`, `git status --porcelain`, `git diff` y
`gh pr checks`. H2 lo documentó con un resumen de «14 pendientes» sobre una propuesta ya terminada
([`../../003-h2-internal-httpx-cliente/gates/evidencia-plataforma.md`](../../003-h2-internal-httpx-cliente/gates/evidencia-plataforma.md)).
**Todo lo que sigue está leído con el paso directo (`rtk proxy`)**, y las salidas son literales.

---

## 1. Publicación de la rama

Antes de empujar, la rama no existía en `origin`: `git ls-remote --heads origin h3-internal-cache-sqlite`
no devolvió ninguna línea.

```
$ git push -u origin h3-internal-cache-sqlite
╭─────────────────────────────────────╮
│ 🥊 lefthook v1.13.6  hook: pre-push │
╰─────────────────────────────────────╯
┃  guardia-push ❯

  ────────────────────────────────────
summary: (done in 0.01 seconds)
✔️ guardia-push (0.01 seconds)
remote:
remote: Create a pull request for 'h3-internal-cache-sqlite' on GitHub by visiting:
remote:      https://github.com/jmorenobl/kitlegal/pull/new/h3-internal-cache-sqlite
remote:
To github.com:jmorenobl/kitlegal.git
 * [new branch]      h3-internal-cache-sqlite -> h3-internal-cache-sqlite
branch 'h3-internal-cache-sqlite' set up to track 'origin/h3-internal-cache-sqlite'.
```

```
$ rtk proxy git rev-parse HEAD origin/h3-internal-cache-sqlite
0b1bf84e40ed3256291ee12e7fd686569d3fdf76
0b1bf84e40ed3256291ee12e7fd686569d3fdf76
$ rtk proxy git rev-list --count main..HEAD
15
```

El gancho `pre-push` (`scripts/lefthook/pre-push/guardia-push.sh`) se ejecutó y dejó pasar la referencia.
Es una rama de hito: no es `main`, no va forzada, no es un borrado ni una etiqueta. `origin` tiene
exactamente el commit local, sin ninguno reescrito.

## 2. Apertura de la propuesta de cambio

No existía propuesta previa para la rama:

```
$ rtk proxy gh pr view h3-internal-cache-sqlite
no pull requests found for branch "h3-internal-cache-sqlite"
```

```
$ gh pr create --base main --head h3-internal-cache-sqlite \
    --title 'feat(H3): internal/cache, caché SQLite con TTL y --offline' \
    --body-file specs/004-h3-internal-cache-sqlite/gates/pr-h3.md
https://github.com/jmorenobl/kitlegal/pull/18
```

El cuerpo, [`pr-h3.md`](./pr-h3.md), sigue la plantilla del ritual (§6.3 del roadmap: objetivo, alcance,
controles añadidos, decisiones y pendientes). Añade tres cosas:

- **La justificación de la dependencia nueva** (constitución §V). `modernc.org/sqlite v1.58.0` está en la
  lista fijada y el hito la nombra (FR-043). La alternativa descartada, `mattn/go-sqlite3`, exige cgo
  (ADR 0002).
- **La diferencia frente a la predicción del supuesto S1.** Los nueve indirectos de la sonda 6 entraron
  exactos. Lo único no previsto es la subida de `golang.org/x/tools` de v0.26.0 a v0.48.0, que pide
  `modernc.org/libc v1.75.6` ([`s1-dependencias.md`](./s1-dependencias.md)).
- **La evidencia adjunta**: [`evidencia-sc013.md`](./evidencia-sc013.md) (los controles fallan ante cada
  intento de saltárselos, T012) y [`evidencia-cierre.md`](./evidencia-cierre.md) (validación agregada,
  T013).

Tras leer los estados, el cuerpo se actualizó con su resultado (§7).

## 3. Integración continua y supuesto S3

```
$ rtk proxy gh run list --branch h3-internal-cache-sqlite
completed	success	feat(H3): internal/cache, caché SQLite con TTL y --offline	ci	h3-internal-cache-sqlite	pull_request	34718912959	6m23s	2026-09-12T21:04:37Z
```

```
$ rtk proxy gh run view 34718912959 --json jobs --jq '.jobs[] | "JOB \(.name)\t\(.conclusion)\t\(.startedAt)\t\(.completedAt)", (.steps[] | "  \(.number)\t\(.name)\t\(.conclusion)\t\(.startedAt)\t\(.completedAt)")'
JOB ci	success	2026-09-12T21:04:40Z	2026-09-12T21:10:59Z
  1	Set up job	success	2026-09-12T21:04:42Z	2026-09-12T21:04:43Z
  2	Obtener el código	success	2026-09-12T21:04:43Z	2026-09-12T21:04:44Z
  3	Instalar Go y restaurar la caché	success	2026-09-12T21:04:44Z	2026-09-12T21:04:56Z
  4	Ejecutar los controles	success	2026-09-12T21:04:56Z	2026-09-12T21:10:43Z
  5	Publicar el perfil de cobertura	success	2026-09-12T21:10:43Z	2026-09-12T21:10:45Z
  9	Post Instalar Go y restaurar la caché	success	2026-09-12T21:10:45Z	2026-09-12T21:10:57Z
  10	Post Obtener el código	success	2026-09-12T21:10:57Z	2026-09-12T21:10:57Z
  11	Complete job	success	2026-09-12T21:10:57Z	2026-09-12T21:10:57Z
```

Un solo flujo (`ci`) con un solo trabajo, en verde. Su paso «Ejecutar los controles» es `make ci`, cuya
línea de prerrequisitos es desde T010 `ci: fmt-check lint test test-integration vuln schema-check secrets
mod-verify mod-tidy-check`.

**`test-integration` está dentro del trabajo.** Lo muestra el registro del run
(`rtk proxy gh run view 34718912959 --log`), sin buscar `PASS` ni `SKIP`: la receta es contrato de H0,
corre sin `-v` y no los imprime. Es la receta ejecutándose entre `test` y `vuln` con el paquete de la caché
en `ok`:

```
ci	Ejecutar los controles	2026-09-12T21:09:18.4822813Z go test -race -tags=integration ./...
ci	Ejecutar los controles	2026-09-12T21:09:21.3033934Z ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	1.016s
ci	Ejecutar los controles	2026-09-12T21:09:21.8499124Z ok  	github.com/jmorenobl/kitlegal/internal	1.369s
ci	Ejecutar los controles	2026-09-12T21:09:24.2488804Z ok  	github.com/jmorenobl/kitlegal/internal/app	2.183s
ci	Ejecutar los controles	2026-09-12T21:09:24.2507912Z ?   	github.com/jmorenobl/kitlegal/internal/app/ejemplo	[no test files]
ci	Ejecutar los controles	2026-09-12T21:09:24.2508964Z ?   	github.com/jmorenobl/kitlegal/internal/app/ejemplo/kitlegal-e2e	[no test files]
ci	Ejecutar los controles	2026-09-12T21:09:29.2688662Z ok  	github.com/jmorenobl/kitlegal/internal/cache	5.086s
ci	Ejecutar los controles	2026-09-12T21:09:29.2689951Z ok  	github.com/jmorenobl/kitlegal/internal/cli	1.511s
ci	Ejecutar los controles	2026-09-12T21:09:29.2691303Z ?   	github.com/jmorenobl/kitlegal/internal/core	[no test files]
ci	Ejecutar los controles	2026-09-12T21:09:29.5579467Z ok  	github.com/jmorenobl/kitlegal/internal/core/schema	1.019s
ci	Ejecutar los controles	2026-09-12T21:10:04.2428724Z ok  	github.com/jmorenobl/kitlegal/internal/httpx	33.126s
ci	Ejecutar los controles	2026-09-12T21:10:04.2430418Z ok  	github.com/jmorenobl/kitlegal/internal/render	1.022s
ci	Ejecutar los controles	2026-09-12T21:10:04.2630155Z go tool -modfile=tools/govulncheck/go.mod govulncheck ./...
[…]
ci	Ejecutar los controles	2026-09-12T21:10:43.6489724Z ci: todos los controles en verde
```

**Veredicto de S3: se cumple.** Los tests de permisos de T010 comprueban una precondición: que el sistema
de ficheros haga valer los permisos. Con la variable `CI` no vacía, que GitHub Actions exporta como
`CI=true`, esa precondición incumplida termina en `t.Fatalf` y no en un salto. Un ejecutor privilegiado
habría dejado `internal/cache` en `FAIL` dentro de `test-integration` y el trabajo `ci` en rojo. El trabajo
está en verde, con la receta ejecutada y `internal/cache` en `ok` a los 5,086 s. Por tanto, el ejecutor de
`ubuntu-latest` no es privilegiado y los tres tests de permisos midieron de verdad.

## 4. Duración de `ci` y supuesto S5

S5 asumía que «`ci` de una PR limpia sigue por debajo de los 3 minutos de H0» con `test-integration`
dentro. El objetivo de H0 es **en caliente**, con la caché de `setup-go` restaurada.

**Esta ejecución fue en frío**:

```
ci	Instalar Go y restaurar la caché	2026-09-12T21:04:56.3521679Z Cache is not found
[…]
ci	Post Instalar Go y restaurar la caché	2026-09-12T21:10:57.4752107Z Cache saved with the key: setup-go-Linux-x64-ubuntu24-go-1.27.1-faad8ae2390159a627abd187a613292da7f4afabaf9cffc8224f62e16a4222ae
```

No es casualidad. La clave de caché de `setup-go` incluye el hash de `go.sum` (`cache-dependency-path` en
`ci.yml`), y H3 cambia `go.sum` al añadir `modernc.org/sqlite`. La primera ejecución de la propuesta que
añade una dependencia es siempre en frío: la caché de `main` tiene otra clave.

Desglose del trabajo (6 m 19 s, de `21:04:40Z` a `21:10:59Z`), con las marcas de tiempo del registro:

| Tramo | Desde → hasta | Duración | Nota |
|---|---|---|---|
| Preparación y `setup-go` | 21:04:40 → 21:04:56 | 16 s | `Cache is not found` |
| `fmt-check` | 21:04:56 → 21:06:42 | 1 m 46 s | descarga y compila `golangci-lint` en frío |
| `lint` | 21:06:42 → 21:07:30 | 48 s | `0 issues.` |
| `test` (`-race`, cobertura) | 21:07:30 → 21:09:18 | 1 m 48 s | incluye compilar `modernc.org/libc` y el driver |
| **`test-integration`** | 21:09:18 → 21:10:04 | **46 s** | la suite entera otra vez con `-tags=integration`; `internal/httpx` 33 s |
| `vuln` | 21:10:04 → 21:10:15 | 11 s | |
| `schema-check` y `secrets` | 21:10:15 → 21:10:36 | 21 s | |
| `mod-verify` y `mod-tidy-check` | 21:10:36 → 21:10:43 | 7 s | |
| Codecov y guardado de la caché | 21:10:43 → 21:10:59 | 16 s | |

**Veredicto de S5: medido en frío, sin medida en caliente en esta tarea.** Estos 6 m 19 s no refutan S5
porque no son la condición que S5 describe, y tampoco lo confirman. Lo que sí queda medido:

- **Coste real de `test-integration` en la integración continua: 46 s**, no los ≈ 22 s que el plan estimó
  a partir de la suite local. La mayor parte es `internal/httpx` repetido bajo `-race` (33 s frente a
  29 s en `test`).
- **Referencia en caliente de H2**: la última ejecución de la propuesta de H2 restauró la caché
  (`Cache restored successfully`, run `34688913176`) y duró ≈ 1 m 50 s de principio a fin del run, sin
  `test-integration`. Sumarle los 46 s medidos aquí da del orden de 2 m 40 s, **una estimación y no una
  medida**: el tiempo en caliente de H3 también cambia por los tests nuevos de `internal/cache`.
- **La medida en caliente llega con el siguiente push de la rama.** El workflow `hito` confirma esta tarea
  y vuelve a publicar la rama, y esa ejecución encontrará la caché que esta guardó
  (`setup-go-Linux-x64-ubuntu24-go-1.27.1-faad8ae…`). Relanzar el trabajo solo para medirlo no es una de
  las órdenes de plataforma que esta tarea tiene autorizadas, y volvería a subir cobertura a Codecov.

## 5. Estados de la propuesta

```
$ rtk proxy gh pr checks h3-internal-cache-sqlite
codecov/patch	fail	0	https://app.codecov.io/gh/jmorenobl/kitlegal/pull/18
ci	pass	6m19s	https://github.com/jmorenobl/kitlegal/actions/runs/34718912959/job/103621019276
codecov/project	pass	1s	https://app.codecov.io/gh/jmorenobl/kitlegal/pull/18
codecov/project/internal/cli	pass	0	https://app.codecov.io/gh/jmorenobl/kitlegal/pull/18
codecov/project/internal/core	pass	0	https://app.codecov.io/gh/jmorenobl/kitlegal/pull/18
```

Código de salida 1, porque un estado está en rojo. Hay cinco estados, ninguno pendiente y ninguno ausente.

El resumen `pass`/`fail` no distingue un verde que midió de un verde vacío («No coverage information
found on base report», `codecov.yml`; H1). Lo distingue el título del check-run:

```
$ rtk proxy gh api repos/jmorenobl/kitlegal/commits/0b1bf84e40ed3256291ee12e7fd686569d3fdf76/check-runs \
    --jq '.check_runs[] | "\(.name)\t\(.status)\t\(.conclusion)\t\(.started_at)\t\(.completed_at)\t\(.output.title)"'
codecov/project/internal/cli	completed	success	2026-09-12T21:10:56Z	2026-09-12T21:10:56Z	98.07% (target 90.00%)
codecov/project/internal/core	completed	success	2026-09-12T21:10:55Z	2026-09-12T21:10:55Z	90.36% (target 85.00%)
codecov/patch	completed	failure	2026-09-12T21:10:54Z	2026-09-12T21:10:54Z	85.64% of diff hit (target 92.47%)
codecov/project	completed	success	2026-09-12T21:10:53Z	2026-09-12T21:10:54Z	90.99% (target 70.00%)
ci	completed	success	2026-09-12T21:04:40Z	2026-09-12T21:10:59Z	null
```

| Estado | Medido | Objetivo | ¿Midió de verdad? | Dónde se fija |
|---|---|---|---|---|
| `codecov/project` | **90,99 %** | 70 % | ✅ porcentaje y objetivo | `codecov.yml`, DoD 9 |
| `codecov/project/internal/core` | **90,36 %** | 85 % | ✅ la base contiene `internal/core/**` desde H1 | `codecov.yml`, DoD 9 |
| `codecov/project/internal/cli` | **98,07 %** | 90 % | ✅ la base contiene `internal/cli/**` desde H1 | `codecov.yml`, aceptación de H1 |
| `codecov/patch` | **85,64 %** del diff | 92,47 % | ✅ midió, y **falla** | **ninguno**: objetivo por omisión de Codecov (§6) |

**Ningún verde vacío.** H3 no declara componente nuevo (supuesto S4, SC-014) y los tres estados de
proyecto traen porcentaje y objetivo. `internal/core` no gana sentencias: el paquete `core` de T001 no las
tiene.

### Cotejo con la medida local

| Umbral | Exigido | Local (`evidencia-cierre.md`) | Codecov |
|---|---|---|---|
| Global | ≥ 70 % | 92,9 % | **90,99 %** |
| `internal/core/**` | ≥ 85 % | 90,1 % | **90,36 %** |
| `internal/cli` | ≥ 90 % | 98,6 % | **98,07 %** |

Las cifras no coinciden al decimal y no tienen por qué. El perfil es el mismo:
`go test -race -shuffle=on -coverprofile=coverage.out ./...`, y el registro remoto da los mismos
porcentajes por paquete que el local (`internal/cache` 87,8 %, `internal/cli` 98,6 %,
`internal/core/schema` 90,1 %). Pero `go tool cover` cuenta **sentencias** y Codecov recuenta el perfil
por **líneas**. Las dos lecturas superan por separado los umbrales de la Definition of Done, y
`codecov.yml` no aparece en el diff frente a `main`.

## 6. Hallazgo: `codecov/patch` en rojo, con un objetivo que ningún documento fija

**Qué es.** `codecov.yml` declara `coverage.status.project` (global, 70 %) y los dos estados de proyecto
de los componentes `internal/core` (85 %) e `internal/cli` (90 %). **No declara ningún estado `patch`.**
Codecov lo emite igual con su configuración por omisión, `target: auto`, que toma como objetivo la
cobertura del informe de la base. El objetivo sube a medida que sube la cobertura de `main`:

| Propuesta | Título de `codecov/patch` |
|---|---|
| H1, #9 | `92.94% of diff hit (target 83.33%)` (`pr-h1.md`) |
| H2, #15 | `94.83% of diff hit (target 89.52%)` (`evidencia-plataforma.md` de H2) |
| **H3, #18** | **`85.64% of diff hit (target 92.47%)`** |

**Por qué no es un incumplimiento de la Definition of Done ni del spec.**

- **Definition of Done, punto 9**: «Cobertura de `internal/core/**` ≥ 85 % y global ≥ 70 % (umbrales en
  `codecov.yml`; no son objetivo, son red de seguridad)». Los dos se cumplen y los dos miden.
- **Spec de H3, SC-014**: `make ci` en verde, cobertura global ≥ 70 % y `internal/core/**` ≥ 85 %. Además,
  en *Fuera de alcance*: «**Umbral de cobertura propio para `internal/cache`**: el hito no fija ninguno;
  rigen los umbrales generales».
- **Roadmap**, entre lo que no se hace: «Cobertura como objetivo: el umbral existe para no retroceder, no
  para perseguirlo».

**De dónde sale el hueco.** Hay que mirar el perfil local, porque el comparador de la API de Codecov
responde `404` sin token en un repositorio privado. Las líneas sin cubrir se concentran en las ramas de
fallo de la apertura y de las migraciones. Una parte la ejercitan **solo** los tests de integración, que
por decisión del plan no alimentan `coverage.out`: la receta `test-integration` es contrato de H0, no
lleva `-coverprofile`, y lo fija la obligación 11 del plan.

```
$ rtk proxy go tool cover -func=coverage.out | rtk proxy grep "internal/cache/"        # solo unitarios, lo que sube ci.yml
[…]abrir.go:88:    creaElFichero            66.7%
[…]abrir.go:153:   conexionDeLectura        77.8%
[…]abrir.go:186:   reabreInmutable           0.0%
[…]abrir.go:251:   dsnSoloLectura           75.0%
[…]abrir.go:264:   cierraTrasElFallo        66.7%
[…]abrir.go:284:   falloAlLeerElEsquema     50.0%
[…]abrir.go:299:   falloAlLeerLoInmutable    0.0%
[…]abrir.go:326:   falloDelContexto          0.0%
[…]migraciones.go:149: aplica               79.2%
[…]migraciones.go:221: falloAlAplicar       66.7%
```

Medida de contraste con un perfil **fuera del repositorio**, en el directorio temporal de la sesión, sumando
los tests de integración. No cambia nada del árbol ni de lo que sube la integración continua:

```
$ rtk proxy go test -race -count=1 -tags=integration -coverprofile=$TMP/cov-cache-integracion.out ./internal/cache/
ok  	github.com/jmorenobl/kitlegal/internal/cache	2.801s	coverage: 91.7% of statements
$ rtk proxy go tool cover -func=$TMP/cov-cache-integracion.out | rtk proxy grep -E "abrir.go|migraciones.go|total"
[…]abrir.go:88:    creaElFichero            83.3%
[…]abrir.go:153:   conexionDeLectura        88.9%
[…]abrir.go:186:   reabreInmutable          73.3%
[…]abrir.go:251:   dsnSoloLectura          100.0%
[…]abrir.go:264:   cierraTrasElFallo        66.7%
[…]abrir.go:284:   falloAlLeerElEsquema     75.0%
[…]abrir.go:299:   falloAlLeerLoInmutable    0.0%
[…]abrir.go:326:   falloDelContexto          0.0%
[…]migraciones.go:149: aplica               79.2%
[…]migraciones.go:221: falloAlAplicar       66.7%
total:                                      91.7%
```

(Rutas abreviadas con `[…]` = `github.com/jmorenobl/kitlegal/internal/cache/`; el resto de funciones del
paquete está en 85,7-100 % en las dos medidas.)

- **Con los tests de integración, `internal/cache` pasa de 87,8 % a 91,7 %.** La reapertura
  `immutable=1` (`reabreInmutable`), que solo se alcanza con un directorio `0500`, sube de 0 % a 73,3 %.
- **Quedan a 0 % en las dos medidas** `falloAlLeerLoInmutable` y `falloDelContexto`: ramas defensivas que
  ningún test, unitario ni de integración, provoca.

**Por qué T014 no lo arregla, y por qué no se decide aquí.**

- T014 solo publica y lee estados. Sus rutas declaradas son cuatro ficheros de `gates/`: no puede tocar
  `internal/cache`, `codecov.yml`, `Makefile` ni `ci.yml`, y el guardián de diff lo rechazaría.
- Cualquiera de los remedios es una decisión sobre qué estados de cobertura son gate. Eso lo fijó H0
  (`codecov.yml`, FR-029: «Los tres estados son bloqueantes […] Sin exclusiones ni excepciones») y lo
  acota el spec de H3 («rigen los umbrales generales»). Es alcance y decisión cerrada: criterio de
  decisión autónoma, punto 4. **Se escala, no se adivina.**
- Nada se rebaja: ningún objetivo cambia y ningún estado se silencia.

**Opciones para la persona que fusiona**, sin orden de preferencia impuesto por esta tarea:

1. **Declarar `patch` en `codecov.yml`** para que sea una decisión explícita y no un valor por omisión que
   se mueve con la base. Puede ser un objetivo fijo o `informational: true`, con su razón escrita, como
   los tres estados actuales. H1 mostró que `codecov/patch` también para código nuevo sin tests (sonda
   #11: `0.00% of diff hit`), así que retirarlo tiene coste.
2. **Tests unitarios para las ramas defensivas** de `abrir.go` y `migraciones.go` que hoy nadie provoca
   (`falloDelContexto`, `falloAlLeerLoInmutable`, `cierraTrasElFallo`, `falloAlAplicar`), en una tarea
   nueva de H3 antes de fusionar. Sube el diff cubierto sin tocar ningún umbral, pero no alcanza lo que
   solo se ejercita con permisos reales.
3. **Subir también la cobertura de los tests de integración**, lo que exige cambiar la receta
   `test-integration` (contrato de H0) o `ci.yml`, y reabrir la obligación 11 del plan.

`main` no tiene protección de rama que impida fusionar con un estado en rojo, así que la decisión es
consciente o no se toma.

## 7. Actualización del cuerpo de la propuesta

El apartado «Pendientes» de [`pr-h3.md`](./pr-h3.md) anunciaba S3 y S5 como pendientes de comprobar. Tras
leer los estados, pasa a llevar el resultado de §3 a §6, incluido el rojo de `codecov/patch` con sus
opciones, para que quien revise la propuesta lo vea sin abrir este fichero. La propuesta se actualizó con
el mismo fichero:

```
$ gh pr edit h3-internal-cache-sqlite --body-file specs/004-h3-internal-cache-sqlite/gates/pr-h3.md
```

---

## Veredicto

| Criterio de T014 | Estado |
|---|---|
| Rama del hito publicada en `origin`, con el gancho `pre-push` ejecutado | ✅ |
| Propuesta de cambio abierta hacia `main` con la plantilla del ritual y la justificación §V | ✅ #18 |
| Evidencia de SC-013 y de cierre adjunta en el cuerpo; diferencia frente a S1 anotada | ✅ |
| Estados de la integración continua y de Codecov leídos con `gh pr checks` | ✅ cinco estados, ninguno ausente ni vacío |
| **S3**: trabajo `ci` en verde con `test-integration` dentro | ✅ |
| **S5**: duración de `ci` de la propuesta anotada | ✅ anotada: 6 m 19 s en frío; en caliente, con el siguiente push |
| Umbrales de la Definition of Done en remoto (global, `internal/core`) sin rebajar | ✅ |
| Sin fusión, sin push a `main`, sin `--force`, sin etiquetas | ✅ |

**Para la persona que revisa y fusiona**:

- `codecov/patch` en rojo (§6): decidir entre las opciones.
- La revisión del ritual (§6.4 del roadmap: `/code-review` y `/security-review`).
- El squash-merge.

Ninguna de las tres es del workflow.

## `make ci` final

Se ejecutó al terminar la tarea, con `pr-h3.md` en su versión final y este fichero escrito salvo esta
sección, así que `gitleaks` los recorrió. T014 no toca ninguna línea de código. Sus únicos ficheros son
los de `gates/` y la marca en `tasks.md`, de modo que el resultado tiene que coincidir con el del cierre, y
coincide.

```
$ rtk proxy make ci > $TMP/make-ci-t014.log 2>&1        # código de salida 0
go tool -modfile=tools/golangci-lint/go.mod golangci-lint fmt --diff ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
go tool -modfile=tools/golangci-lint/go.mod golangci-lint run ./...
# github.com/golangci/golangci-lint/v2/cmd/golangci-lint
ld: warning: -bind_at_load is deprecated on macOS
0 issues.
go test -race -shuffle=on -coverprofile=coverage.out ./...
ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	1.648s	coverage: 0.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal	1.629s	coverage: [no statements]
ok  	github.com/jmorenobl/kitlegal/internal/app	3.898s	coverage: 93.1% of statements
	github.com/jmorenobl/kitlegal/internal/app/ejemplo		coverage: 0.0% of statements
	github.com/jmorenobl/kitlegal/internal/app/ejemplo/kitlegal-e2e		coverage: 0.0% of statements
ok  	github.com/jmorenobl/kitlegal/internal/cache	3.196s	coverage: 87.8% of statements
ok  	github.com/jmorenobl/kitlegal/internal/cli	2.225s	coverage: 98.6% of statements
?   	github.com/jmorenobl/kitlegal/internal/core	[no test files]
ok  	github.com/jmorenobl/kitlegal/internal/core/schema	2.428s	coverage: 90.1% of statements
ok  	github.com/jmorenobl/kitlegal/internal/httpx	16.293s	coverage: 96.8% of statements
ok  	github.com/jmorenobl/kitlegal/internal/render	3.346s	coverage: 95.8% of statements
go test -race -tags=integration ./...
ok  	github.com/jmorenobl/kitlegal/cmd/kitlegal	(cached)
ok  	github.com/jmorenobl/kitlegal/internal	(cached)
ok  	github.com/jmorenobl/kitlegal/internal/app	(cached)
?   	github.com/jmorenobl/kitlegal/internal/app/ejemplo	[no test files]
?   	github.com/jmorenobl/kitlegal/internal/app/ejemplo/kitlegal-e2e	[no test files]
ok  	github.com/jmorenobl/kitlegal/internal/cache	(cached)
ok  	github.com/jmorenobl/kitlegal/internal/cli	(cached)
?   	github.com/jmorenobl/kitlegal/internal/core	[no test files]
ok  	github.com/jmorenobl/kitlegal/internal/core/schema	(cached)
ok  	github.com/jmorenobl/kitlegal/internal/httpx	(cached)
ok  	github.com/jmorenobl/kitlegal/internal/render	(cached)
go tool -modfile=tools/govulncheck/go.mod govulncheck ./...
No vulnerabilities found.
schema-check: no hay schemas/ todavía; los aportan H4 (borrador) y H10 (contrato)
go tool -modfile=tools/gitleaks/go.mod gitleaks dir . --redact --no-banner
11:18PM INF scanned ~8475453 bytes (8.48 MB) in 819ms
11:18PM INF no leaks found
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

(Los códigos de color de `gitleaks` se han quitado de las dos líneas `INF`.)

- **`test-integration` sale `(cached)`**, y es legítimo: T014 no cambia ningún fichero Go ni material de
  test desde `0b1bf84`. La ejecución sin caché de esa misma suite está en `evidencia-cierre.md` §1, y la
  remota en §3 de este fichero.
- **Las cifras de cobertura** son las mismas que en `evidencia-cierre.md`.

```
$ rtk proxy git status --porcelain --untracked-files=all
 M specs/004-h3-internal-cache-sqlite/gates/tarea-actual.json
 M specs/004-h3-internal-cache-sqlite/gates/tareas-intentos.json
?? specs/004-h3-internal-cache-sqlite/gates/evidencia-plataforma.md
?? specs/004-h3-internal-cache-sqlite/gates/pr-h3.md
```

**Ficheros que toca T014**: los dos que el workflow actualiza al lanzarla (`tarea-actual.json` y
`tareas-intentos.json`); `gates/pr-h3.md`, que es el cuerpo de la propuesta; este fichero; y la marca de
T014 en `tasks.md`. Ningún fichero de producto, nada bajo `testdata/` ni `schemas/`. **T014 queda marcada**:
`make ci` en verde y todos sus criterios cumplidos. El rojo de `codecov/patch` se escala a decisión humana
en §6, sin ocultarlo ni rebajar nada.
