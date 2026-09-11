# T021 · intento 1 (2026-09-11) — **sin marcar**, bloqueada por un prerrequisito humano

**Estado de la tarea**: `[ ]`. **`gates/pr-h1.md` no se ha creado**: no existe todavía ninguna propuesta de
cambio de H1, así que no hay resultado, ni duración, ni enlace que registrar. Crear el fichero con
apartados «pendiente» habría dado cuerpo de evidencia a algo que nadie ha medido.

## La causa, exacta

T021 mide **tres cosas que solo la plataforma produce**: el estado `codecov/project/internal/cli` sobre una
propuesta de cambio, la duración de la ejecución de `ci.yml` y el enlace a esa propuesta. Ninguna de las
tres es observable desde este árbol, y la cadena que lo impide está cortada en el primer eslabón:

1. **La rama del hito no está en el remoto.**

   ```console
   $ git ls-remote --heads origin
   b99349ea…  refs/heads/enfoque-skills-territorio
   573f4e87…  refs/heads/main            # h1-kernel-cli-multicall no está

   $ git rev-parse --abbrev-ref --symbolic-full-name @{u}
   fatal: no upstream configured for branch 'h1-kernel-cli-multicall'
   ```

2. **Subirla está denegado por política del propio repositorio**, no por un fallo. `.claude/settings.json`
   lista `Bash(git push:*)` y `Bash(git merge:*)` en `permissions.deny`, y el intento quedó rechazado:

   ```console
   $ git push -u origin h1-kernel-cli-multicall
   Permission to use Bash with command git push … has been denied.
   ```

   Es una decisión deliberada y vigente —el workflow `hito` reserva abrir la propuesta y el squash-merge a
   la persona (`revision_humana_final`: «approve = listo para abrir PR y squash-merge (acción humana)»)—,
   así que **no se busca un rodeo**. `curl` y `wget` también están denegados; no hay ninguna otra vía, ni
   debe haberla.

3. **Sin propuesta de cambio no hay ejecución que medir.** `ci.yml` solo se dispara con `pull_request` y
   con `push` a `main`:

   ```yaml
   on:
     pull_request:
     push:
       branches: [main]
   ```

   Y sin ejecución no hay subida a Codecov, luego no hay estado `codecov/project/internal/cli` que pueda
   aparecer ni bloquear. `gh pr checks` no tiene sobre qué operar.

**Lo que esto NO es**: no es el umbral. La obligación 5 del plan advierte que, si el estado no apareciera,
la causa sería la configuración de Codecov y nunca el umbral. Aquí no es ni una cosa ni la otra: el estado
no ha tenido ocasión de calcularse. **El objetivo del 90 % no se ha tocado y no debe tocarse.**

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

El estado a comprobar se llama **`codecov/project/internal/cli`** —Codecov nombra el estado con el campo
`name` del componente, no con `component_id`—, igual que en H0 apareció `codecov/project/internal/core`
(`gates/verificacion-pr.md` de H0). Debe salir con `bucket` distinto de informativo: `informational: false`
está declarado.

## Lo que sí quedó comprobado en este intento, y no sustituye a lo anterior

Es material de apoyo para el intento que sí pueda ejecutarse; **ninguna de estas medidas vale como
evidencia de T021**, porque T021 mide la plataforma.

**`make ci` en verde en local**, con los ocho controles de H0 más lo que añade H1:

```console
$ make ci
0 issues.                                        # golangci-lint (fmt --diff y run, con TESTDATA_PKGS)
ok  github.com/jmorenobl/kitlegal/internal/app          3.186s  coverage: 90.8% of statements
ok  github.com/jmorenobl/kitlegal/internal/cli          2.625s  coverage: 98.3% of statements
ok  github.com/jmorenobl/kitlegal/internal/core/schema  2.219s  coverage: 90.0% of statements
ok  github.com/jmorenobl/kitlegal/internal/render       2.813s  coverage: 95.8% of statements
No vulnerabilities found.                        # govulncheck
no leaks found                                   # gitleaks
all modules verified                             # go mod verify, raíz y las cuatro herramientas
ci: todos los controles en verde
```

`internal/cli` al **98,3 %** deja **8,3 puntos** de margen sobre el 90 % que exige `codecov.yml`. El estado
de la plataforma no mide exactamente lo mismo (Codecov cuenta líneas del componente sobre el informe
subido, no sentencias por paquete), pero un margen así hace improbable que el motivo de un eventual fallo
sea la cobertura real y no la configuración —que es justo la disyuntiva que la obligación 5 manda resolver.

**Exclusiones: ninguna nueva sin justificar** (FR-057). El diff de configuración contra `main` se reduce a
tres ficheros —`.golangci.yml`, `codecov.yml` y `Makefile`— y `.github/` no cambia en absoluto:

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

**Duración**: `make ci` completo, con el e2e dentro y caché tibia, **12 s** frente al techo de 180 s que
fijó H0 (T020, `gates/quickstart-h1.md`, obligación 4). Es la medida **local**; la del flujo de la
plataforma sigue sin tomar y es la que T021 debe registrar.

## Lo que no se hizo, a propósito

- **No se subió la rama ni se abrió la propuesta.** La política del repositorio lo deniega y el workflow lo
  reserva a la persona.
- **No se rebajó el objetivo del 90 %** ni ningún otro umbral de `codecov.yml`, ni se tocó
  `informational: false`.
- **No se creó `gates/pr-h1.md`** con resultados inventados, estimados ni «pendientes».
- **No se marcó T021 como `[X]`.** Hacerlo afirmaría que SC-004 y SC-013 están validados en la plataforma
  sin estarlo, que es el mismo error que los intentos 1-3 de T014 en H0 evitaron cometer.
- **No se tocó ningún fichero fuera de esta nota**, que está dentro del directorio del hito.
