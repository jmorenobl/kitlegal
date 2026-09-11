# T021 · intento 2 (2026-09-11) — **sin marcar**, el mismo prerrequisito humano, re-verificado

**Estado de la tarea**: `[ ]`. **`gates/pr-h1.md` sigue sin crearse**: no existe todavía ninguna propuesta de
cambio de H1, así que no hay resultado, ni duración de la plataforma, ni enlace que registrar. Crear el
fichero con apartados «pendiente» daría cuerpo de evidencia a algo que nadie ha medido.

Este intento **re-verifica** que el bloqueo sigue en pie (no se dio por supuesto desde el intento 1) y añade
lo único que sí se puede adelantar desde este árbol: el **pre-vuelo de la configuración de cobertura**, que
es exactamente la disyuntiva que manda resolver la obligación 5 —«si no apareciera, la causa es la
configuración de la plataforma, nunca el umbral»—.

## La causa, re-verificada hoy (2026-09-11 18:43Z, commit `c2a6d63`)

T021 mide **tres cosas que solo la plataforma produce**: el estado `codecov/project/internal/cli` sobre una
propuesta de cambio, la duración de la ejecución de `ci.yml` y el enlace a esa propuesta. La cadena que lo
impide sigue cortada en el primer eslabón:

1. **La rama del hito no está en el remoto.** Medido de nuevo, no heredado:

   ```console
   $ git ls-remote --heads origin
   b99349ea…  refs/heads/enfoque-skills-territorio
   573f4e87…  refs/heads/main            # h1-kernel-cli-multicall sigue sin estar

   $ git rev-parse --abbrev-ref --symbolic-full-name @{u}
   fatal: no upstream configured for branch 'h1-kernel-cli-multicall'
   ```

   La lectura del remoto **sí funciona** (`ls-remote` responde), así que esto no es un fallo de red ni de
   credenciales: la rama no se ha subido.

2. **Subirla está denegado por política del propio repositorio.** `.claude/settings.json` mantiene
   `Bash(git push:*)` y `Bash(git merge:*)` en `permissions.deny`, junto con `curl` y `wget`. Es una
   decisión deliberada y vigente —el workflow `hito` reserva abrir la propuesta y el squash-merge a la
   persona (`revision_humana_final`: «approve = listo para abrir PR y squash-merge (acción humana)»)—, así
   que **no se busca un rodeo**.

3. **Sin propuesta de cambio no hay ejecución que medir.** `ci.yml` solo se dispara con `pull_request` y con
   `push` a `main`; subir la rama, por sí solo, no dispararía nada. Sin ejecución no hay subida a Codecov,
   luego no hay estado `codecov/project/internal/cli` que pueda aparecer ni bloquear.

4. **Añadido de este intento**: `gh` tampoco está disponible aquí. No está denegado por política, pero cada
   invocación (`gh pr list`, `gh auth status`) pide aprobación interactiva y la sesión es desatendida. Es
   secundario: aunque estuviera concedido, no habría propuesta sobre la que operar.

**Lo que esto NO es**: no es el umbral. Aquí el estado no ha tenido ocasión de calcularse. **El objetivo del
90 % no se ha tocado y no debe tocarse.**

## Pre-vuelo de la configuración de cobertura (lo nuevo de este intento)

La obligación 5 anticipa un fallo concreto —que el estado no llegue a aparecer— y prohíbe atribuirlo al
umbral. Lo que puede comprobarse **sin plataforma** ya está comprobado, para que el intento que sí se
ejecute no tenga que descubrirlo:

| Comprobación | Resultado |
|---|---|
| Nombre que tendrá el estado | `codecov/project/internal/cli`. Codecov nombra el estado con el campo `name` del componente, no con `component_id`; en H0 salió `codecov/project/internal/core` (`001-h0…/gates/verificacion-pr.md`, l. 119) |
| Carácter bloqueante | `informational: false` declarado en el componente, como los tres estados de H0 |
| Umbrales de H0 | intactos: 70 % de proyecto y 85 % de `internal_core` sin tocar |
| Fichero de informe que sube `ci.yml` | `./coverage.out`, el que genera `make test` (`go test -race -shuffle=on -coverprofile=coverage.out ./...`), con `fail_ci_if_error: true` |
| ¿Hay líneas de `internal/cli` en ese informe? | Sí: **200 bloques** de **7 ficheros** (`describe.go`, `errors.go`, `globales.go`, `log.go`, `parse.go`, `preescaneo.go`, `sobre.go`) |
| `presentador.go` no aparece en el informe | **Correcto, y no resta**: es una interfaz sin sentencias ejecutables, así que no aporta líneas al informe ni al componente |
| Cobertura del paquete | **98,3 %** de sentencias, **8,3 puntos** por encima del objetivo |

**El único punto del pre-vuelo que no se puede cerrar en local** —y por tanto el primer sitio donde mirar si
el estado no apareciera— es el arreglo de rutas. El perfil de Go nombra los ficheros con el prefijo del
módulo:

```
github.com/jmorenobl/kitlegal/internal/cli/describe.go:79.2,80.16 2 15
```

mientras que `paths: internal/cli/**` se resuelve contra rutas relativas a la raíz del repositorio. La
plataforma normaliza esas rutas cotejando el informe con la lista de ficheros del repositorio, y por eso el
componente debería casar tal cual está declarado. Si aun así el estado saliera ausente o a 0 %, **la causa
es esa normalización y el remedio documentado es una entrada `fixes:` en `codecov.yml`
(`github.com/jmorenobl/kitlegal/::`), nunca rebajar el 90 %.** H0 no despeja esta duda: allí el componente
`internal_core` no casaba con ningún fichero, y lo que quedó demostrado es otra cosa igual de útil —que la
plataforma honra el `codecov.yml` **de la rama de la propuesta**, no el de `main`, porque el componente se
declaró en la rama de H0 y su estado salió en aquella propuesta antes de integrar nada—.

## Qué hace falta para desbloquearla (prerrequisito humano)

Una persona con permiso de push:

```console
$ git push -u origin h1-kernel-cli-multicall
$ gh pr create --base main --head h1-kernel-cli-multicall --title 'feat(H1): kernel de la línea de órdenes y multicall'
```

Y después, sobre esa propuesta, lo que T021 debe leer y registrar en `gates/pr-h1.md`:

```console
$ gh pr checks h1-kernel-cli-multicall --json name,state,bucket
$ gh run list --branch h1-kernel-cli-multicall --json databaseId,conclusion,startedAt,updatedAt
```

Criterio de aceptación, sin margen de interpretación:

- `codecov/project/internal/cli` **aparece** en la lista y con `bucket` distinto de informativo. `PENDING`
  colgado no vale: en una rama con controles obligatorios equivale a un bloqueo permanente.
- `ci` termina en `SUCCESS` con los ocho controles de H0 más los que añade H1.
- La duración se toma de `startedAt`→`updatedAt` de la ejecución, y se contrasta con el techo de 3 minutos
  de SC-006 que fijó H0 (allí: 2 min 48 s en frío, 1 min 00 s en caliente).

## Material de apoyo, que **no** sustituye a lo anterior

Medido hoy sobre `c2a6d63`; **ninguna de estas medidas vale como evidencia de T021**, porque T021 mide la
plataforma.

**`make ci` en verde en local**, con los controles de H0 más lo que añade H1:

```console
$ make ci
0 issues.                                        # golangci-lint (fmt --diff y run, con TESTDATA_PKGS)
ok  github.com/jmorenobl/kitlegal/internal/app          2.473s  coverage: 90.8% of statements
ok  github.com/jmorenobl/kitlegal/internal/cli          1.892s  coverage: 98.3% of statements
ok  github.com/jmorenobl/kitlegal/internal/core/schema  1.979s  coverage: 90.0% of statements
ok  github.com/jmorenobl/kitlegal/internal/render       2.132s  coverage: 95.8% of statements
No vulnerabilities found.                        # govulncheck
no leaks found                                   # gitleaks
all modules verified                             # go mod verify, raíz y las cuatro herramientas
ci: todos los controles en verde
```

**Duración local**: 11,4 s con caché tibia, frente al techo de 180 s. Es la medida **local**; la del flujo
de la plataforma sigue sin tomar y es la que T021 debe registrar.

**Exclusiones: ninguna nueva sin justificar** (FR-057). El diff de configuración contra `main` se reduce a
tres ficheros —`.golangci.yml`, `codecov.yml` y `Makefile`— y **`.github/` no cambia en absoluto**
(`git diff --stat main...HEAD -- .github/` vacío):

| Cambio | Signo |
|---|---|
| `errcheck.exclude-functions` (`fmt.Fprint`/`Fprintf`/`Fprintln`) | **retirada**: el control se endurece |
| `forbidigo` `path-except: ^internal/` de H0 | **sustituida** por cuatro exclusiones acotadas a `path` **y** `text`, sobre las dos raíces de composición (`^cmd/` y `^internal/app/testdata/kitlegal-e2e/`) y solo para las marcas `R4:` y `R5-descriptores:` |
| `depguard` con R1, R2 y R3 | control **nuevo** |
| `forbidigo` con `analyze-types: true` y R4/R5 por símbolo | control **nuevo**, más estricto |
| `codecov.yml` | componente `internal_cli` **añadido** (90 %, `informational: false`); umbrales de H0 (70 % global, 85 % `internal_core`) **intactos** |
| `Makefile` | `TESTDATA_PKGS` **amplía** lint y formato a `internal/app/testdata/*`; `test-e2e` deja de ser un aviso y ejecuta los guiones |

Las cuatro exclusiones de `forbidigo` llevan su justificación escrita en el propio `.golangci.yml` y la
obligación 8 del plan las ejerció en T020: la misma violación que pasa en `kitlegal-e2e` **falla** en
`internal/app/testdata/ejemplo`, y `fmt.Println` falla en las dos rutas.

## Lo que no se hizo, a propósito

- **No se subió la rama ni se abrió la propuesta.** La política del repositorio lo deniega y el workflow lo
  reserva a la persona.
- **No se rebajó el objetivo del 90 %** ni ningún otro umbral de `codecov.yml`, ni se tocó
  `informational: false`. Tampoco se añadió `fixes:` «por si acaso»: es el remedio a un fallo que aún no se
  ha observado, y añadirlo a ciegas es tocar la configuración sin diagnóstico.
- **No se creó `gates/pr-h1.md`** con resultados inventados, estimados ni «pendientes».
- **No se marcó T021 como `[X]`.** Hacerlo afirmaría que SC-004 y SC-013 están validados en la plataforma
  sin estarlo, que es el mismo error que los intentos 1-3 de T014 en H0 evitaron cometer —y que el intento 4
  de aquella tarea solo pudo cerrar **después** de que la persona resolviera su prerrequisito—.
- **No se tocó ningún fichero fuera de esta nota**, que está dentro del directorio del hito.

---

# T021 · intento 1 (2026-09-11) — sin marcar

Mismo diagnóstico, con el mismo detalle: rama sin subir, `git push` denegado por `permissions.deny`, `ci.yml`
sin disparador que alcanzar y, por tanto, ningún estado de cobertura que pudiera aparecer ni bloquear. El
intento 2 lo re-verificó en vez de darlo por bueno, y añadió el pre-vuelo de la configuración de Codecov de
más arriba. No se creó `gates/pr-h1.md` ni se marcó la tarea.
