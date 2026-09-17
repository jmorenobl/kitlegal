# Revisión final de H5.1: pendiente para una persona

> **Resuelto** (2026-09-17). La repetición por etiqueta se hizo: la primera, sobre `064308f`, dio `fallo` por una traza
> ilegible que arregló T015 (`73fe6d2`); la segunda, sobre `73fe6d2`, dio `aprobado` (`gates/evals-aceptacion.md`,
> sección vigente). Lo que sigue es la nota tal como la dejó la corrección de la ronda 1.

Anotado por la corrección de la ronda 1 (2026-09-16). Ningún motivo de los jueces exigió tocar `testdata/` ni
`schemas/`, ni pidió una decisión humana: los cuatro motivos del juez A (criterio f) están corregidos en `7ae7a05`
(documentación y comentarios) y en el commit de `specs/` que lo registra. Queda una acción de plataforma que el
corrector no puede hacer, porque no empuja ni etiqueta.

## Repetir la ejecución de aceptación (FR-070, SC-007)

La aceptación registrada en `gates/evals-aceptacion.md` —la ejecución de apertura
[35148840549](https://github.com/jmorenobl/kitlegal/actions/runs/35148840549), sobre `8272bc8`— **deja de cubrir la
cabeza**: `7ae7a05` cambia, fuera de `specs/007-h5-1-avisos-de-vigencia/`, estos ficheros:

- `CHANGELOG.md`
- `CONTRIBUTING.md`
- `README.md`
- `internal/evals/conjunto.go`
- `internal/evals/formato.go`

Por el contrato de la ejecución de aceptación §6, se repite por etiqueta, como última acción de plataforma, después
de que la revisión final quede aprobada por los dos jueces y de que se empuje la rama (el push no lanza el job de
evals: no tiene `synchronize`):

1. Actualizar el cuerpo de la propuesta de cambio #34 con el fichero corregido:
   `gh pr edit 007-h5-1-avisos-de-vigencia --body-file specs/007-h5-1-avisos-de-vigencia/gates/pr-h5.1.md`.
2. Quitar la etiqueta `evals` si está puesta y volver a ponerla, e identificar la ejecución nueva por el último evento
   `labeled`: `quickstart.md` §11.6.
3. Esperarla y leerla con la tercera orden de §11.3 y las de §11.4, y comprobarla con las dos órdenes de §11.5, con las
   mismas condiciones: veredicto `aprobado`, `red` vacío, la tasa de `18-lrjpac-norma-derogada.yaml` con el modelo que
   decide sobre 3 sesiones y el reparto de sus dos avisos en cada sesión, y ningún fichero fuera del directorio del
   hito entre el commit evaluado y la cabeza.
4. Añadir su evidencia como sección nueva y vigente de `gates/evals-aceptacion.md`.

Si una ronda posterior de la revisión vuelve a cambiar algo fuera de `specs/007-h5-1-avisos-de-vigencia/`, la
repetición va detrás del último de esos cambios.
