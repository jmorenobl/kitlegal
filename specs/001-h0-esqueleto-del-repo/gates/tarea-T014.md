# T014 · no completada: prerrequisito humano ausente

**Fecha**: 2026-09-10 · **Intento**: 1 de 3 · **Estado de la tarea**: `[ ]` (sin marcar, a propósito)
**Verificación determinista**: `make ci` en **verde** (todos los controles, cobertura 83.3 %).

## Por qué queda sin marcar aun con `make ci` en verde

`make ci` verifica el árbol, no el objetivo de esta tarea. T014 existe para medir el veredicto **de la
plataforma** sobre dos propuestas de cambio (SC-005 y SC-006), y **el repositorio no tiene remoto**:

```console
$ git remote -v
            # sin salida
```

Sin `origin` no hay push, sin push no hay propuesta de cambio, y `ci.yml` —que se dispara por
`pull_request` y por push a `main`— no se ejecuta. Los cuatro resultados y las dos duraciones que la tarea
debe registrar **no son obtenibles**, y rellenarlos con mediciones locales sería medir otra cosa.

Marcar `[X]` afirmaría que SC-005 y SC-006 están validados. No lo están. La tarea misma prohíbe simular
resultados, así que se deja `[ ]`.

## Qué sí se entregó

`gates/verificacion-pr.md`, con:

1. El bloqueo documentado y los seis huecos declarados **PENDIENTE**/**SIN MEDIR**, ninguno inventado.
2. Los dos prerrequisitos humanos que lo desbloquean: crear el repositorio en la plataforma y darlo de alta
   como `origin`, y dar de alta el secreto `CODECOV_TOKEN`.
3. La mitad local del control que el escenario mide, sí ejecutada y en verde: `forbidigo` bloquea
   `fmt.Println` bajo `internal/**` y lo permite bajo `cmd/**` (FR-014), con fixture creado y borrado.
4. Una **discrepancia real** encontrada al hacerlo: el fixture que `quickstart.md` dicta literalmente para
   el escenario 11 no produce el hallazgo de `forbidigo` que su apartado «Esperado» anuncia, porque el
   procesador `uniq-by-line` (activo por omisión) deja un solo hallazgo por línea y el `exported` de
   `revive` ocupa la misma línea 5. Aislado por bisección y confirmado con `uniq-by-line: false`. El
   control no está roto y no se tocó `.golangci.yml`; lo que falla es el guion del escenario, cuya
   corrección está fuera de las rutas declaradas por T014.

## Qué no se hizo, a propósito

No se creó el repositorio remoto (acción hacia fuera y decisión de alcance, clasificada por la propia tarea
como prerrequisito humano), no se rebajó `fail_ci_if_error` ni ninguna otra bandera, y no se modificó
`.golangci.yml`.

## Para el siguiente intento

No hay nada que reintentar en el repositorio: los intentos 2 y 3 fallarían por lo mismo mientras no exista
el remoto. El desbloqueo es humano. Una vez exista `origin` y esté el secreto, el escenario 11 se ejecuta
tal cual está escrito y se rellenan las seis casillas de `gates/verificacion-pr.md` —con el fixture
corregido según el punto 4, o leyendo el hallazgo concreto y exigiendo que sea `forbidigo`.
