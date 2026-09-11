# Nota de T009 — `Descripcion` y el diccionario de `misspell`

**Estado de T009: verde.** Esta nota no describe un fallo; describe una deuda que conviene saldar en una
tarea que sí declare `.golangci.yml` entre sus rutas (T012 y T016 lo hacen).

## Qué ocurre

El contrato del applet (`contracts/registro-y-describe.md` §1 y `data-model.md` §6) fija el identificador
`Descripcion` en `Applet.Descripcion()` y en `Verbo.Descripcion`. El linter `misspell`, cuyo diccionario es
solo inglés (`locale: US`), lo lee como una errata de `Description` y falla. Es el mismo falso positivo que
H0 neutralizó para `argumentos` con `misspell.ignore-rules` en `.golangci.yml` (research.md D9).

## Qué se hizo en T009, y por qué

T009 no declara `.golangci.yml` entre sus rutas, así que el guardián de diff rechazaría el cambio de
configuración. Las cuatro apariciones del identificador llevan un `//nolint:misspell` **específico y con
explicación** —lo que `nolintlint` exige y lo único que quedaba dentro de las rutas de la tarea—:

- `internal/app/applet.go`: el método de la interfaz y el campo del struct.
- `internal/app/registro_test.go`: la implementación del método y el literal del verbo de prueba.

El resto de falsos positivos en español (`asume`, `distribuye`, `producto`, `calcular`, `previos`) se
resolvieron reescribiendo el texto, sin `nolint`.

## Qué conviene hacer

Añadir `descripcion` a `misspell.ignore-rules` en `.golangci.yml` —junto a `argumentos`, que ya está— y
retirar los cuatro `//nolint:misspell`. Sin eso, **cada applet que se escriba a partir de H4 necesitará su
propia excepción**, porque todos implementan `Descripcion()`: la excepción por línea no escala y la regla
de ignorado sí. Retirar `nolint` endurece el control y no es una exclusión nueva de las que FR-057 prohíbe.
