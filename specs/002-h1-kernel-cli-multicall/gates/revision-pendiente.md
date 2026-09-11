# Revisión final de H1 — pendientes que requieren a una persona

**Fecha**: 2026-09-11 · **Ronda**: corrección tras `gates/revision-a.json` (aprobado) y
`gates/revision-b.json` (rechazado, 13 motivos) · **Rama**: `h1-kernel-cli-multicall`

Los trece motivos del juez B están corregidos en la rama (commits `fix(H1)`, `test(H1)` y `docs(H1)`
posteriores a `461331f`) y `make ci` termina en verde, **salvo tres puntos que el corrector no puede
cerrar por sí solo**. Ninguno bloquea `make ci`; los tres son acciones de la persona.

## 1. `//nolint:misspell` por `Descripcion` — toca `internal/app/testdata/` (motivo [h])

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
