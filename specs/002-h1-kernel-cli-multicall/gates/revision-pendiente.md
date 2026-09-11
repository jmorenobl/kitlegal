# Revisión final de H1 — pendientes que requieren a una persona

## Ronda 2 · 2026-09-12 · A rechazado (5 motivos) / B aprobado

**Ronda**: corrección tras `gates/revision-a.json` (rechazado, criterios b, d, e y f) y
`gates/revision-b.json` (aprobado) · **Rama**: `h1-kernel-cli-multicall`.

Los cinco motivos del juez A están corregidos en la rama y `make ci` termina en verde. **No queda
ninguna decisión pendiente de esta ronda**; lo que sigue es el rastro de cada corrección y una decisión
de diseño que conviene conocer.

| Motivo | Corrección | Dónde |
|---|---|---|
| [e][f] La tubería cerrada mataba el proceso por `SIGPIPE` en lugar de terminar con `1` | La raíz de composición del kernel desarma la señal al arrancar (`signal.Ignore(syscall.SIGPIPE)`), de modo que la escritura devuelve `EPIPE` y sigue el camino ya escrito: presentador → traducción única → código `1` y mensaje en la salida de error. Test nuevo sobre el binario real con una tubería cuyo lector se cierra antes de arrancarlo, para el sobre, la tabla, las tres ayudas, `version` y `--describe` | `internal/app/main.go` (`desarmarTuberiaCerrada`), `internal/app/tuberia_test.go` (`TestTuberiaCerrada`); research.md D15, ADR 0006, CHANGELOG, contratos de banderas y del sobre, quickstart esc. 5, plan.md fila 11 |
| [d][b] Nada validaba la salida real de los applets de ejemplo contra el esquema que ellos mismos emiten | Test nuevo que, por cada verbo de cada applet de ejemplo, pide el esquema con `--describe`, lo compila con `AssertFormat()` y valida contra él un sobre de éxito real y dos de fallo reales; comprueba además que la validación restringe `data` (una clave de más o de menos deja de validar). La tabla exige una fila por verbo registrado, así que un verbo nuevo sin caso falla en lugar de quedar sin validar. La mutación del juez (`Salida: cuenta{}` en `echo`) lo hace fallar | `internal/app/esquema_test.go` (`TestSalidaContraSuEsquema`); plan.md filas 4 y 5 |
| [b] `TestCodigoSalida/correcto` y `TestAppletHereda/{echo,contar}` dependían de `KITLEGAL_LOG` | Las dos tablas fijan la variable —en el test y en cada subcaso— y dejan de ser paralelas, con el mismo patrón y la misma explicación que `TestSalidaEstandarRota`, `TestDryRun` y `TestPuntoDeEntrada`. Comprobado: `KITLEGAL_LOG=debug` y `KITLEGAL_LOG=verbose` (inválido) en el entorno dejan `internal/app`, `internal/cli` y `cmd/...` en verde | `internal/app/codigos_test.go`, `internal/app/hereda_test.go` |
| [f] ADR 0005 afirmaba que `--describe` «no puede mentir» sobre `data` | La consecuencia se reformula a lo que es cierto —no puede mentir sobre la forma **declarada**— y el punto ciego pasa a «En contra, y asumido»: `Salida` y `Resultado.Datos` son dos declaraciones que nada relaciona en compilación, y la coincidencia la vigila el test anterior, que es el control de la DoD §1.4 | `docs/ADR/0005-contrato-de-applet.md` |
| [f] El contrato del sobre decía que el pre-escaneo reconoce solo las tres formas largas | Añadido el token exacto `-h`, con las mismas exclusiones que research.md D25 regla 1 | `contracts/sobre-de-salida.md` §5 |

**Decisión de diseño que conviene conocer.** La señal se desarma en `app.Main` y no en los dos
`package main`, por dos razones: la garantía es del kernel y no de un binario concreto, de modo que todo
ejecutable que compone el kernel la hereda sin poder olvidarla; y el `main` del binario de e2e vive bajo
`internal/app/testdata/`, material protegido que el corrector tiene prohibido tocar, y sin él el test
con la tubería real no podría ejercer el binario que el e2e construye. Alternativas y motivo en
research.md D15.

**Observaciones del juez B que no eran motivos** y quedan como están, por si la persona quiere
resolverlas: `gates/quickstart-h1.md:307` conserva la cifra de 18 `//nolint` de la ejecución de T020
(es el registro de aquella ejecución; hoy son dos, `internal/arch_test.go` y
`internal/app/tuberia_test.go`, las dos `gosec` G204 sobre un subproceso con argumentos literales y
justificadas en la propia línea); y `spec.md` FR-026 nombra solo `--help` mientras `-h` está en los
contratos, en research.md D25 y en el CHANGELOG. El spec está cerrado por su gate, así que anotarlo
allí es decisión de la persona.

Sigue vigente de la ronda 1 el punto 2 (copiar a la propuesta #9 la tabla de módulos) y, como
información, el 3.

## Ronda 1 · 2026-09-11 · A aprobado / B rechazado (13 motivos)

**Ronda**: corrección tras `gates/revision-a.json` (aprobado) y `gates/revision-b.json` (rechazado, 13
motivos) · **Rama**: `h1-kernel-cli-multicall`

Los trece motivos del juez B están corregidos en la rama (commits `fix(H1)`, `test(H1)` y `docs(H1)`
posteriores a `461331f`) y `make ci` termina en verde, **salvo tres puntos que el corrector no puede
cerrar por sí solo**. Ninguno bloquea `make ci`; los tres son acciones de la persona.

## 1. `//nolint:misspell` por `Descripcion` — toca `internal/app/testdata/` (motivo [h]) — **resuelto**

> **Resuelto por la persona** tras esta corrección: parche aplicado sin cambios, `make ci` en verde y
> commit `refactor(H1)` en la rama. No queda ninguna `//nolint:misspell`; el parche se ha retirado del
> repositorio. Lo que sigue documenta por qué no lo aplicó el corrector.

**Qué pide el motivo.** Añadir `descripcion` a `misspell.ignore-rules` en `.golangci.yml` y retirar las
directivas `//nolint:misspell` que neutralizan el identificador `Descripcion` (fijado por el contrato del
applet), cinco de ellas en los applets de ejemplo, que T014 exigía «sin ninguna excepción de lint».

**Por qué no está hecho.** Las dos mitades son inseparables: en cuanto la regla de ignorado entra,
`nolintlint` rechaza como *unused* toda directiva que quede, y cinco de ellas viven en
`internal/app/testdata/ejemplo/echo.go` y `contar.go`. Modificar ficheros existentes bajo `testdata/` es
capa 3 de la constitución («Gates», humano) y el corrector tiene prohibido tocarlos. Retirar solo las
directivas de fuera de `testdata/` dejaría `make ci` en rojo, así que no se ha hecho ninguna de las dos
mitades: la rama sigue con las directivas (18 tras esta corrección: las 16 anteriores más dos en el
applet de prueba nuevo de `internal/app/main_test.go`, escritas con la misma justificación).

**Qué hay preparado.** [`revision-pendiente-misspell.patch`](./revision-pendiente-misspell.patch): la
regla de ignorado, la retirada de las 18 directivas y el realineado de `gofumpt` que deja la retirada,
generado sobre una copia del árbol fuera del repositorio. Sobre esa copia, `golangci-lint run` sobre
`./...` y los dos paquetes de `testdata` termina en **0 issues** y los tests de `internal/app` y
`cmd/kitlegal` en verde; `git apply --check` confirma que aplica limpio sobre la rama.

**Qué tiene que hacer la persona.** Desde la raíz del repositorio:

```sh
git apply specs/002-h1-kernel-cli-multicall/gates/revision-pendiente-misspell.patch
make ci
git commit -am "refactor(H1): descripcion en misspell.ignore-rules; retira los nolint por Descripcion"
```

Después, este parche puede borrarse. Si se prefiere no tocar `testdata/` en este hito, la deuda queda
documentada en `gates/nota-T009-misspell.md` y `gates/nota-T014-verificacion.md` y el motivo [h] sigue
abierto: es una decisión, no un olvido.

## 2. La descripción de la propuesta #9 en la plataforma (motivo [i], FR-060)

La constitución §V exige justificar «en el plan (sección Complexity Tracking) y en la PR» todo módulo
fuera de la lista. Los cuatro que `invopop/jsonschema` arrastra al binario están justificados ya en
`plan.md` y en `gates/pr-h1.md` («Dependencias que el binario enlaza»). **Falta copiar esa tabla a la
descripción de la propuesta #9** en GitHub: editar la propuesta es una acción sobre la plataforma que la
política del repositorio reserva a la persona, igual que el push. La tabla está lista para pegar en
`gates/pr-h1.md`.

## 3. Las dos entradas de `.gitignore` retiradas (motivo [i], menor)

`.gitignore` vuelve a la versión de `main`: las entradas `.specify/integrations/.cache/` y
`.claude/worktrees/` eran ajenas al spec y a cualquier tarea de H1, como dice el motivo. Consecuencia
operativa en esta máquina, que conviene resolver **antes de lanzar el siguiente hito**:

- `.specify/integrations/.cache/` es una caché de catálogos que spec-kit regenera al arrancar el
  workflow. El corrector la ha borrado (cuatro ficheros JSON, 28 KB, regenerables) para dejar el árbol
  limpio, pero volverá a aparecer, y el workflow hace `git add -A` en sus pasos de commit: sin una
  exclusión, la caché acabaría versionada (el mismo mecanismo por el que un worktree acabó como gitlink
  en la PR #8).
- El corrector no ha podido escribir en `.git/info/exclude` (fichero protegido por la política de la
  sesión).

> **Hecho en local**: la persona ha añadido las dos entradas a `.git/info/exclude` de esta máquina.
> Queda llevarlas a `main` en un commit de mantenimiento propio antes de lanzar H2.

**Decisión de la persona**: o bien traer las dos entradas a `main` en un commit de mantenimiento propio
—es la opción limpia: son hygiene del repositorio, no del hito—, o bien excluirlas solo en local con
`printf '.specify/integrations/.cache/\n.claude/worktrees/\n' >> .git/info/exclude`.

## Sin decisión pendiente, solo información

- `go.yaml.in/yaml/v4` entra en el binario en versión candidata (`v4.0.0-rc.2`), fijada por
  `pb33f/ordered-map/v2`; no existe todavía una `v4` estable y Dependabot la seguirá. Justificado en
  `plan.md` y vigilado por `TestDependenciasDelBinario`.
- El veredicto de la ronda fue A aprobado / B rechazado; según `workflow.yml` ese desacuerdo escala a
  humano y no pasa por el corrector. Esta corrección se ha ejecutado a petición, aplicando los motivos de
  B; los dos JSON de los jueces no se han tocado.
