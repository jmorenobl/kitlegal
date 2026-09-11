# Evidencia en la propuesta de cambio · T021 (SC-004 · SC-013 · FR-057 · obligación 5)

**Fecha**: 2026-09-11 · **Hito**: H1 · **Tarea**: T021 · **Rama**: `h1-kernel-cli-multicall`
**Propuesta de cambio**: [#9](https://github.com/jmorenobl/kitlegal/pull/9) contra `main`, commit `3783f2d`
**Sondas**: [#10](https://github.com/jmorenobl/kitlegal/pull/10) (limpia) y
[#11](https://github.com/jmorenobl/kitlegal/pull/11) (sucia), ambas contra `h1-kernel-cli-multicall`,
**cerradas sin integrar**
**Intentos**: 1, 2 y 3 (desatendidos) — bloqueados: la rama no estaba en el remoto y `git push` está
denegado por política · **4 (este, sesión interactiva)** — la persona subió la rama y abrió las propuestas.

## Veredicto

**Ejecutado y en verde, con un hallazgo sobre la plataforma que obligó a medir con sondas.** En la
propuesta del hito, `codecov/project/internal/cli` aparece pero **no mide**: sale `SUCCESS` con «No
coverage information found on base report», un verde vacío. Contra la rama del hito, que ya tiene informe
con `internal/cli`, el mismo estado **mide 97,77 %** sobre el código real y **falla con 81,98 %** en cuanto
se añade código sin test. **SC-004, SC-013 y FR-057 quedan validados; el objetivo del 90 % no se tocó.**

## Resultados

| Qué | Criterio | Resultado |
|---|---|---|
| Flujo `ci` en la propuesta del hito | SC-013 | **VERDE** · #9, ejecución 34638125765, `SUCCESS` en sus dos intentos |
| Controles de H0 más los de H1 | SC-013, FR-057 | **VERDE** · los ocho de `make ci`, con `TESTDATA_PKGS` en el lint, el e2e en `internal/app` y el test de arquitectura en `internal` (detalle abajo) |
| Subida de cobertura | research.md D15 (H0) | **VERDE** · `Upload queued for processing complete`, token enmascarado, `fail_ci_if_error` intacto |
| `codecov/project/internal/cli` aparece | obligación 5 | **Aparece** en #9, #10 y #11, `informational: false` |
| …y **mide** el kernel | SC-004 | **97,77 % (target 90,00 %)** · #10, `SUCCESS`. En local: 98,3 % de sentencias, 96,82 % de líneas |
| …y **bloquea** | obligación 5 | **81,98 % (target 90,00 %)** · #11, `FAILURE`, con `ci` en `SUCCESS`: lo que bloquea es el componente |
| Duración **en caliente** | < 3 min (SC-006 de H0, obligación 4) | **1 min 11 s** (job `19:23:30Z` → `19:24:41Z`, intento 2 de 34638125765, `Cache hit`) |
| Duración **en frío** | se registra, sin umbral | **3 min 19 s** (job `19:18:36Z` → `19:21:55Z`, intento 1, `Cache is not found`) |
| Exclusiones nuevas | FR-057 | **Ninguna sin justificar** (tabla abajo) |

El caliente se queda por debajo de la mitad del techo, así que la contingencia de la obligación 4 (reutilizar
el binario del e2e o sacarlo a un job propio) no hace falta. El frío sube frente a H0 (2 min 48 s con go1.26.6;
3 min 05 s ya con go1.27.1 en la PR de H0): son las cinco dependencias nuevas y la compilación del binario
del e2e. No tiene umbral.

## El hallazgo: en la propuesta que introduce un componente, su estado es un verde vacío

En #9 (base `main`) los estados de Codecov fueron:

```console
$ gh api repos/jmorenobl/kitlegal/commits/3783f2d…/check-runs --jq '.check_runs[]|{name,conclusion,title:.output.title}'
{"name":"codecov/project/internal/cli",  "conclusion":"success","title":"No coverage information found on base report"}
{"name":"codecov/project/internal/core", "conclusion":"success","title":"No coverage information found on base report"}
{"name":"codecov/patch",                 "conclusion":"success","title":"92.94% of diff hit (target 83.33%)"}
{"name":"codecov/project",               "conclusion":"success","title":"92.94% (target 70.00%)"}
{"name":"ci",                            "conclusion":"success","title":null}
```

**La causa está en el código de Codecov, no en nuestra configuración.** En
[`apps/worker/services/notification/notifiers/mixins/status.py`](https://github.com/codecov/umbrella/blob/main/apps/worker/services/notification/notifiers/mixins/status.py),
`_get_project_status` hace tres comprobaciones en orden antes de comparar con el objetivo:

1. sin cobertura en el head → `if_not_found` con «No coverage information found on head»;
2. sin informe en la base → `if_not_found` con «No report found to compare against»;
3. informe en la base **sin cobertura** → `if_not_found` con «No coverage information found on base report».

Solo si pasa las tres llega a `_get_target` y a comparar con el objetivo, **aunque el objetivo sea fijo**.
En un componente, head y base se filtran por sus `paths`. `main` sí tiene informe (cada push a `main` sube
cobertura, y por eso `codecov/project` mide 92,94 %), pero no tiene ni un fichero en `internal/cli/` ni en
`internal/core/`: la base filtrada no tiene cobertura, se aplica la regla 3 y `if_not_found` vale `success`
por omisión. **En la propuesta que crea el árbol de un componente, su estado no puede fallar, sea cual sea
la cobertura.**

Tres consecuencias:

- **No es un problema de rutas y `fixes:` no hace falta.** El mensaje no es el de la regla 1, así que el head
  filtrado por `internal/cli/**` sí tiene cobertura: el prefijo del módulo (`github.com/jmorenobl/kitlegal/…`)
  se normaliza bien. Las sondas lo confirman con el porcentaje.
- **El «verde» de `internal/core` en H0 también era vacío.** `verificacion-pr.md` de H0 lo dio por
  contestado («lo resuelve en verde»); lo que Codecov hizo fue saltarse la evaluación. Desde H1 `internal/core`
  mide de verdad: 89,47 % frente al 85 % en #10 y #11.
- **El comentario de `codecov.yml` se queda corto.** Dice que un umbral declarado antes de que exista el código
  «empieza a aplicarse por sí solo en cuanto ese árbol tenga ficheros». Es cierto a partir de la **siguiente**
  propuesta, cuando la base ya tiene el árbol, pero no en la que lo introduce. Queda como pendiente (abajo):
  T021 no declara `codecov.yml`.

## Las sondas: medir y bloquear contra una base que sí tiene el componente

Dos ramas desechables desde `3783f2d`, en `git worktree` fuera del repositorio (en el scratchpad de la
sesión), con propuestas en borrador **contra `h1-kernel-cli-multicall`**. Esa rama ya tiene informe de
Codecov con `internal/cli`, subido por la ejecución de #9.

| Sonda | Commit | Contenido | `make ci` local | `codecov/project/internal/cli` | `codecov/patch` |
|---|---|---|---|---|---|
| #10 limpia | `d88f910` | commit vacío (`--allow-empty`) | — | **`SUCCESS` · 97.77% (target 90.00%)** | `SUCCESS` · coverage not affected |
| #11 sucia | `ac99bed` | `internal/cli/sonda.go`: una función exportada con 50 sentencias sin test | verde, `internal/cli` al 83,8 % | **`FAILURE` · 81.98% (target 90.00%)** | `FAILURE` · 0.00% of diff hit |

En las dos, `ci` terminó en `SUCCESS` (#10: 34639836823; #11: 34639841707). La sonda sucia pasa todos los
controles del repositorio, así que **el único control que la para es el estado del componente**: eso es lo
que la obligación 5 pedía ver. Las dos ejecuciones fueron en frío (`Cache is not found`), porque la caché
guardada en la ejecución de #9 pertenece a la referencia de esa propuesta y no se comparte con otra. No se
usan para medir duración.

`gh pr checks` de las dos:

```console
$ gh pr checks 10 --json name,state,bucket
[{"bucket":"pass","name":"codecov/project/internal/cli","state":"SUCCESS"},
 {"bucket":"pass","name":"codecov/project/internal/core","state":"SUCCESS"},
 {"bucket":"pass","name":"codecov/patch","state":"SUCCESS"},
 {"bucket":"pass","name":"codecov/project","state":"SUCCESS"},
 {"bucket":"pass","name":"ci","state":"SUCCESS"}]

$ gh pr checks 11 --json name,state,bucket
[{"bucket":"fail","name":"codecov/project/internal/cli","state":"FAILURE"},
 {"bucket":"pass","name":"codecov/project/internal/core","state":"SUCCESS"},
 {"bucket":"fail","name":"codecov/patch","state":"FAILURE"},
 {"bucket":"pass","name":"codecov/project","state":"SUCCESS"},
 {"bucket":"pass","name":"ci","state":"SUCCESS"}]
```

Que `codecov/project` global siga en `SUCCESS` en #11 (86,47 % frente al 70 %) enseña para qué sirve el
componente: con el umbral global solo, esa caída habría entrado sin que nada lo impidiera.

## La propuesta del hito: los controles, uno por uno

Ejecución 34638125765 sobre `3783f2d`, `event: pull_request`. El paso «Ejecutar los controles» es `make ci`
y el log muestra:

```
golangci-lint fmt --diff ./... ./internal/app/testdata/ejemplo ./internal/app/testdata/kitlegal-e2e
golangci-lint run ./... ./internal/app/testdata/ejemplo ./internal/app/testdata/kitlegal-e2e
0 issues.
go test -race -shuffle=on -coverprofile=coverage.out ./...
ok  github.com/jmorenobl/kitlegal/cmd/kitlegal            coverage: 0.0% of statements
ok  github.com/jmorenobl/kitlegal/internal                coverage: [no statements]   # TestArquitectura
ok  github.com/jmorenobl/kitlegal/internal/app            coverage: 90.8% of statements  # e2e con testscript
ok  github.com/jmorenobl/kitlegal/internal/cli            coverage: 98.3% of statements
ok  github.com/jmorenobl/kitlegal/internal/core/schema    coverage: 90.0% of statements
ok  github.com/jmorenobl/kitlegal/internal/render         coverage: 95.8% of statements
No vulnerabilities found.
schema-check: no hay schemas/ todavía; los aportan H4 (borrador) y H11 (contrato)
no leaks found
all modules verified            # raíz y las cuatro herramientas
go mod tidy -diff
ci: todos los controles en verde
```

Son los ocho controles de H0: formato, lint (con `gosec` como SAST), tests con detector de carreras,
vulnerabilidades, esquemas, secretos, verificación de módulos y dependencias saneadas. Y lo que añade H1
dentro de ellos: `depguard` y `forbidigo` con R1–R5, los paquetes de `internal/app/testdata` en el lint y
el formato, el e2e de `testscript` y `TestArquitectura`. Las mismas cifras de cobertura por paquete que en
local.

## Exclusiones: ninguna nueva sin justificar (FR-057)

`git diff --name-only main...HEAD` limitado a configuración devuelve `.golangci.yml`, `Makefile` y
`codecov.yml`; **`.github/` no cambia**.

| Cambio | Signo |
|---|---|
| `errcheck.exclude-functions` (`fmt.Fprint`/`Fprintf`/`Fprintln`) | **retirada**: el control se endurece |
| `forbidigo` `path-except: ^internal/` de H0 | **sustituida** por cuatro exclusiones acotadas a `path` **y** `text`, sobre las dos raíces de composición (`^cmd/` y `^internal/app/testdata/kitlegal-e2e/`), solo para las marcas `R4:` y `R5-descriptores:` y justificadas en el propio fichero |
| `depguard` con R1, R2 y R3 | control **nuevo** |
| `forbidigo` con `analyze-types: true` y R4/R5 por símbolo | control **nuevo**, más estricto |
| `codecov.yml` | componente `internal_cli` **añadido** (90 %, `informational: false`); 70 % global y 85 % de `internal_core` **intactos** |
| `Makefile` | `TESTDATA_PKGS` **amplía** lint y formato; `test-e2e` deja de ser un aviso |

Directivas en el código: **17**, todas con su justificación en la misma línea. Son 16 `//nolint:misspell`
por el identificador `Descripcion` que fija el contrato (el diccionario de `misspell` solo es inglés; la deuda
está en `nota-T009-misspell.md`) y 1 `//nolint:gosec` en `internal/arch_test.go` (G204: el ejecutable es
constante y los argumentos son literales y rutas del propio árbol).

## Cierre y limpieza

```console
$ gh pr list --state all --json number,state,headRefName,baseRefName
#11 prueba-cli-sucia  → h1-kernel-cli-multicall  CLOSED
#10 prueba-cli-limpia → h1-kernel-cli-multicall  CLOSED
#9  h1-kernel-cli-multicall → main               OPEN

$ git worktree list
/Users/jorge/Projects/kitlegal  3783f2d [h1-kernel-cli-multicall]   # los dos temporales, eliminados
```

Las dos sondas se cerraron **sin integrar**, con un comentario que da su resultado. Los worktrees y las ramas
locales están borrados. **Las ramas remotas `prueba-cli-limpia` y `prueba-cli-sucia` siguen en `origin`**:
borrarlas es un `git push --delete`, que la política del repositorio reserva a la persona.

## Pendientes que deja este registro

- **Documentar el verde vacío donde se configura.** El comentario de `codecov.yml` debería decir que el
  umbral de un componente no se evalúa en la propuesta que crea su árbol, solo en las siguientes, y
  `verificacion-pr.md` de H0 debería remitir aquí. Lo sensato es hacerlo en el mismo cambio que declare el
  siguiente componente.
- **Todo hito que declare un componente nuevo tendrá que medirlo con una sonda contra su propia rama**, como
  aquí. Si no, su criterio de cobertura quedará «validado» por un verde vacío.
- **Borrar las ramas remotas de las sondas** (acción humana, arriba).

## Lo que no se hizo, a propósito

- **No se rebajó el 90 %** ni ningún otro umbral, y no se tocó `informational: false` ni `if_not_found`.
  Poner `if_not_found: failure` en el componente habría hecho fallar la #9 por falta de base, no por falta de
  cobertura.
- **No se añadió `fixes:`**: el diagnóstico descarta el problema de rutas.
- **No se integró ninguna sonda**, y `sonda.go` no llegó nunca a la rama del hito.
- **No se inventó ningún número.** Todos salen de `gh pr checks`, de las `check-runs` de Codecov y de
  `startedAt`/`completedAt` del job de cada ejecución.
