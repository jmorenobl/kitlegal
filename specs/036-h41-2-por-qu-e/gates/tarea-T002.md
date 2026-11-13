# T002, intento 1: sin verde por una ruta que la tarea no declaraba

## Qué quedó hecho (en el árbol de trabajo)

- `internal/evals/activaciones.go`, nuevo: `Activacion{Skill, Antes}` (claves `skill`, `antes`), `usoDelTranscript`,
  `activacionesDeLosUsos(usos, esDeKitlegal) ([]Activacion, error)` (nunca nil), `activaAntes`, `skillDelUso`,
  `reconocedorDeKitlegal{herramientas, interprete}` con `esDeKitlegal(uso, modo)` (`switch`: `ModoHerramienta`,
  `ModoOrden`, `default: false`) y `consultaKitlegal`, `nuevoReconocedorDeKitlegal()` (devuelve sin envolver los errores
  de `herramientasDelRegistro()` y `nuevoInterprete()`, como hoy `tablaDelLaboratorio`), `leerLosUsos` sin la exigencia
  de `message.id`, `usosDelMensaje`, `bloqueConTipo`, `conResultado`. Se comparó lo quitado de
  `laboratorio_tabla.go` con el fichero nuevo (`git diff --no-index`): fuera del receptor `c` → `r`, el `default` y el
  error retirado, el código es el mismo.
- `internal/evals/laboratorio_tabla.go`: `contadorDelLaboratorio{familia, reconocedor, filas}`; `tablaDelLaboratorio`
  construye el reconocedor con `nuevoReconocedorDeKitlegal()`; `anotarLosUsos` llama a
  `c.reconocedor.esDeKitlegal(uso, modo)`. Nada más cambia.
- `internal/evals/laboratorio_tabla_test.go`: `TestOrdenesDeBash` construye `reconocedorDeKitlegal{herramientas,
  interprete}` en lugar del contador; ningún valor esperado cambia.
- `internal/evals/activaciones_test.go`, nuevo: `TestActivacionesDeLosUsos` (12 casos en los modos `herramienta`,
  `orden` y `""`, con los de la tarea más una herramienta sin prefijo, una orden `kitlegal` en el modo herramienta,
  una herramienta del registro en el modo `""`, ningún uso y un `Skill` sin su entrada) y `TestLaActivacionEnJSON`
  (las claves `skill` y `antes`; la lista vacía se escribe `[]`). Reutiliza `skillNormativa` (conjunto_test.go) y
  `skillDeConsultaJuridica` (informe_test.go), que ya existían en el paquete.

Visto en esta sesión:

- Rojo primero: sin el código, el test no compilaba (`undefined: Activacion`, `activacionesDeLosUsos`,
  `nuevoReconocedorDeKitlegal`).
- Verde: `TestActivacionesDeLosUsos`, `TestLaActivacionEnJSON`, `TestTablaDeLasFamilias`, `TestTablaDeNormativa`,
  `TestTablaDelLaboratorio`, `TestLecturaConCd`, `TestReglaTestigo`, `TestOrdenesDeBash`; las suites congeladas
  `TestEntregaDelHito/h39-describe`, `h40-laboratorio-tabla`, `h41-laboratorio-tabla`, `h41-skills-instaladas` y
  `h41-1-skills-instaladas`.
- Los tres mutantes, en rojo y revertidos: `activa <= primera` → falla «en el mismo turno que la primera herramienta y
  en uno posterior»; `case ModoOrden, "":` → falla «una orden kitlegal en una eval sin binario ni servidor»;
  `var activaciones []Activacion` → fallan «ninguna activación», «ningún uso» y `TestLaActivacionEnJSON`.
- En verde, uno a uno, todos los objetivos de `make ci` salvo `lint`: `fmt-check`, `test`, `test-integration`,
  `test-tiempos`, `vuln` («No vulnerabilities found.»), `schema-check`, `skills-check`, `goreleaser-check`, `secrets`
  («no leaks found»), `mod-verify` y `mod-tidy-check`. (El `make -k` de todos juntos pasó del límite de 10 minutos de
  la herramienta, así que se corrieron por grupos.)

## Por qué no quedó en verde

`make ci` falla en `lint`, y solo por `misspell`:

```
internal/evals/activaciones.go:10: `Activacion` is a misspelling of `Activation` (misspell)
internal/evals/activaciones.go:13: `Activacion` is a misspelling of `Activation` (misspell)
internal/evals/activaciones.go:32: `Activacion` is a misspelling of `Activation` (misspell)
internal/evals/activaciones.go:43: `Activacion` is a misspelling of `Activation` (misspell)
internal/evals/activaciones.go:55: `Activacion` is a misspelling of `Activation` (misspell)
internal/evals/activaciones_test.go:60: … (y 12 líneas más del test, todas `Activacion`)
18 issues:
* misspell: 18
make: *** [lint] Error 1
```

El nombre `Activacion` lo fijan la tarea, data-model §1 y contracts/informe.md, y lo usan después T004 (`juzgar.go`,
`informe.go`) y T008 (`laboratorio_tabla.go`): no se puede cambiar. El diccionario de `misspell` es solo inglés, y la
convención del proyecto (`.golangci.yml:294-295`: «el español se neutraliza falso positivo a falso positivo en
ignore-rules») es una entrada en `linters.settings.misspell.ignore-rules`; el repositorio no tiene ningún
`//nolint:misspell`. `.golangci.yml` no estaba entre las rutas de la tarea, y el guardián de diff habría rechazado el
intento si se tocaba; por eso la línea de T002 en tasks.md se ha ampliado con esa ruta.

## Qué hace el intento siguiente

Añadir a `.golangci.yml`, como primera entrada de `ignore-rules` (orden alfabético, antes de `administrativo`), con un
comentario como los de las demás:

```yaml
        # «Activacion», el tipo de una skill activada de H41.2 (data-model §1 de
        # H41.2); misspell lo lee como «Activation».
        - activacion
```

Comprobado en esta sesión sobre una copia del fichero fuera del repositorio (en `$TMPDIR`, con
`run.relative-path-mode: wd` para que las exclusiones por ruta se resuelvan desde el repositorio):
`go tool -modfile=tools/golangci-lint/go.mod golangci-lint run --config <copia> ./...` da `0 issues.` Con esa entrada
y el código de arriba, `make ci` debería quedar en verde; no se ha podido medir con el fichero del repositorio. Si el
árbol de trabajo no conserva el código de este intento, se rehace como dice la línea de T002.

# T002, intento 2: en verde

El árbol conservaba el código del intento 1; este intento solo añade a `.golangci.yml` la entrada `activacion`, primera
de `linters.settings.misspell.ignore-rules`, con su comentario. Visto en esta sesión, sobre el árbol final:

- `make lint` con el fichero del repositorio: `0 issues.`
- Los tres mutantes, otra vez en rojo y revertidos: `activa <= primera` → falla «en el mismo turno que la primera
  herramienta y en uno posterior»; `case ModoOrden, "":` → falla «una orden kitlegal en una eval sin binario ni
  servidor»; `var activaciones []Activacion` → fallan «ninguna activación», «ningún uso» y `TestLaActivacionEnJSON`.
- Los doce objetivos de `make ci`, en verde: `fmt-check`, `lint`, `test`, `test-integration`, `test-tiempos`, `vuln`,
  `schema-check`, `skills-check`, `goreleaser-check`, `secrets`, `mod-verify` y `mod-tidy-check`. Se corrieron en cuatro
  órdenes y no en un solo `make ci`: solo `make test` tarda 4 min 54 s, y todos juntos pasan del límite de 10 minutos
  por orden de la herramienta. El `make ci` de una pieza no se ha medido en esta sesión.
- Ningún test ni guion espera el error retirado «el mensaje assistant no tiene message.id».

# T002, reparación tras la verificación: `vuln` en rojo por el parche de Go, no por la tarea

La verificación determinista (`gates/ci.log`) cayó en `make vuln`, con todo lo demás en verde: `govulncheck` da
GO-2026-6859 («HTTP/1 server crash due to MaxBytesReader header race in net/http»), encontrada en
`net/http@go1.27.3` y corregida en `net/http@go1.27.4`. En el intento 2, horas antes, `make vuln` decía «No
vulnerabilities found.» sobre el mismo árbol: el hallazgo aparece por el paso del tiempo (la base de `vuln.go.dev`), no
por el código de T002, y deja en rojo a cualquier tarea del run. Es el caso previsto en
`.github/workflows/nightly.yml:11-16` y en `CONTRIBUTING.md`, «Subir el parche de Go (directiva `toolchain`)»: el
arreglo es subir la directiva, no rebajar el control. Precedente: `bab5e085` (`chore(deps): go1.27.3`, #180), en H26.1.

Hecho, y nada más:

- `go.mod`: `toolchain go1.27.3` → `toolchain go1.27.4`. La directiva `go` no cambia; el `Makefile` la lee y exporta
  `GOTOOLCHAIN`. `go.mod` es una ruta que el guardián admite en toda tarea sin declararla
  (`scripts/workflow/guardian-diff.sh:8` y `:72`), así que la línea de T002 no se redelimita.
- `CHANGELOG.md` (también admitida por el guardián): en `[Unreleased]`, una sección «Seguridad» con la subida, como
  la de go1.27.3.

Descartado: dejar la tarea en `[ ]` a la espera de una propuesta aparte sobre `main` (el paso cuarto del procedimiento
manual de `CONTRIBUTING.md`): ninguna tarea del run saldría en verde hasta entonces y la subida va igualmente
versionada y a la vista en la propuesta del hito; y cualquier forma de callar el hallazgo, que es rebajar el control.

Consecuencia que ve el cierre: `internal/evals/eleccion.go:279` elige todas las evals cuando cambia `toolchain` entre
la base y la cabeza, de modo que las medidas de `main` no se reutilizan en el cierre de este hito. No se ha medido en
esta sesión; se lee en el código.

Visto en esta sesión, con go1.27.4 (`make check-tools` lo obtuvo sin error) y con el `go.mod` final; la entrada de
`CHANGELOG.md` y estas notas se escribieron después, y tras ellas solo se repitieron `secrets` («no leaks found») y
`mod-tidy-check`, en verde (ningún fichero Go lee `CHANGELOG.md`):

- Los doce objetivos de `make ci`, en verde: `fmt-check`, `lint` (`0 issues.`), `mod-verify`, `mod-tidy-check`,
  `schema-check` y `skills-check`; `test` (`internal/evals`, 303 s); `test-integration` (`internal/evals`, 352 s);
  `test-tiempos`, `vuln` («No vulnerabilities found.», dos veces), `goreleaser-check` y `secrets` («no leaks found»).
  En cinco órdenes y no en un solo `make ci`: `test` y `test-integration` juntos pasan del límite de 10 minutos por
  orden de la herramienta. El `make ci` de una pieza no se ha medido en esta sesión.
- `scripts/workflow/guardian-diff.sh tarea` no se ha podido ejecutar en esta sesión (la orden pide aprobación); que
  admite `go.mod` y `CHANGELOG.md` se ha leído en su código, no medido.
- Los valores `go1.27.3` y `go1.27.4` de `internal/evals/eleccion_test.go` son de un árbol escrito en el test, no los
  de `go.mod`: no cambian.
