# T014 · no completada: prerrequisito humano ausente

**Fecha**: 2026-09-10 · **Intento**: 2 de 3 · **Estado de la tarea**: `[ ]` (sin marcar, a propósito)
**Verificación determinista**: `make ci` en **verde** (todos los controles, cobertura 83.3 %).

## Por qué sigue sin marcarse

`make ci` verifica el árbol, no el objetivo de esta tarea. T014 existe para medir el veredicto **de la
plataforma** sobre dos propuestas de cambio (SC-005 y SC-006), y el bloqueo del intento 1 sigue vigente,
reverificado al empezar este intento:

```console
$ git remote -v
            # sin salida
$ git branch -a
* h0-esqueleto-del-repo
  main        # ninguna rama remota
```

Sin `origin` no hay push, sin push no hay propuesta de cambio, y `ci.yml` —que se dispara por
`pull_request` y por push a `main`— no se ejecuta. Los cuatro resultados y las dos duraciones que la tarea
debe registrar **no son obtenibles**, y rellenarlos con mediciones locales sería medir otra cosa.

Marcar `[X]` afirmaría que SC-005 y SC-006 están validados. No lo están. La tarea prohíbe simular
resultados, así que se deja `[ ]`.

## Qué aporta este intento respecto al 1

El intento 1 encontró que **el fixture que el guion del escenario 11 dicta no produce el hallazgo de
`forbidigo` que su propio apartado «Esperado» anuncia** —el procesador `uniq-by-line`, activo por omisión,
deja un solo hallazgo por línea, y el `exported` de `revive` ocupa la misma línea 5 y lo desplaza—. Tuvo que
dejarlo anotado sin arreglar porque `quickstart.md` no figuraba entre sus rutas declaradas.

**El intento 2 sí declara `specs/001-h0-esqueleto-del-repo/quickstart.md`, y lo corrige.** Rojo → verde
reproducido en local, sin red y sin tocar `.golangci.yml`:

| Fixture | `make lint` |
|---|---|
| El literal del guion (intento 1) | `revive: 2` — **ningún** `forbidigo` |
| El corregido, ya en `quickstart.md` | `forbidigo: 1` — el hallazgo que FR-014 exige |

Cambios en el guion: el `printf` lleva ahora comentario de paquete y de función exportada (la escritura
prohibida queda como único defecto), el paso pasa a exigir `forbidigo: 1` en lugar de un fallo genérico, y
el «Esperado» incorpora la advertencia de que **no basta con que el lint falle: hay que leer qué regla
falló**. Sin esto, un escenario 11 ejecutado tal cual habría dado SC-005 por bueno dejando FR-014 sin
comprobar nunca.

El fixture se borró al terminar: `internal/prueba/` inexistente y `git status --porcelain -- internal/`
vacío.

## Qué no se hizo, a propósito

No se creó el repositorio remoto (acción hacia fuera y decisión de alcance —visibilidad, propietario,
protecciones de rama—, clasificada por la propia tarea como prerrequisito humano), no se rebajó
`fail_ci_if_error` ni ninguna otra bandera, y no se modificó `.golangci.yml`: el defecto estaba en el
fixture del guion, no en la configuración del lint.

## Para el siguiente intento

**No hay nada que reintentar en el repositorio.** El intento 3 fallaría por lo mismo mientras no exista el
remoto; el desbloqueo es humano y consta en `gates/verificacion-pr.md`:

1. crear el repositorio en la plataforma y darlo de alta como `origin`;
2. dar de alta el secreto `CODECOV_TOKEN`.

Hecho eso, el escenario 11 —ya corregido— se ejecuta tal cual está escrito y se rellenan las seis casillas
de `gates/verificacion-pr.md`.
