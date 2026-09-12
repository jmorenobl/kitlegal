# S1 · la dependencia medida frente a la predicción (T004)

Comprobación del supuesto **S1** de research.md D18 en la tarea que añade la dependencia (T004,
obligación 1 del plan). Lo que sigue es lo que `go mod tidy` escribió de verdad, para el cierre del hito
(T013).

## Las tres comprobaciones, en verde

| Comprobación | Resultado |
|---|---|
| `go mod tidy -diff` (dentro de `make ci`, receta `mod-tidy-check`) | limpio |
| `go mod graph`: el driver no arrastra nada que su propio `go.mod` no declare | se cumple (ver abajo) |
| `TestDependenciasDelBinario` **sin tocar su lista** (`modulosDelBinario`) | en verde |

`TestDependenciasDelBinario` sigue pasando con la lista de H2 intacta porque ningún paquete alcanzable
desde `cmd/kitlegal` importa todavía `internal/cache`: la dependencia existe en `go.mod` y no en el
binario distribuido (SC-012, FR-044; el control propio lo añade T011).

## Lo que la predicción acertó

`go.mod` gana **una** dependencia directa, `modernc.org/sqlite v1.58.0`, y exactamente los nueve
indirectos que la sonda 6 había medido: `modernc.org/libc v1.75.6`, `modernc.org/mathutil v1.7.1`,
`modernc.org/memory v1.12.1`, `golang.org/x/sys v0.47.0`, `github.com/dustin/go-humanize v1.0.1`,
`github.com/google/uuid v1.6.0`, `github.com/mattn/go-isatty v0.0.24`,
`github.com/ncruces/go-strftime v1.0.0` y `github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec`.

Tampoco entraron, como predijo la sonda, `modernc.org/fileutil` ni `github.com/google/pprof`: el `go.mod`
del driver los declara, pero ningún paquete alcanzable los importa, así que se quedan en nodos del grafo
y fuera de `go.mod`. La subida de `golang.org/x/sys` (v0.26.0 → v0.47.0) que S1 anticipaba también
ocurrió, y en la versión que la sonda había medido.

## La diferencia: `golang.org/x/tools` sube a v0.48.0

Lo único que S1 no había previsto: la selección mínima de versiones **sube `golang.org/x/tools` de
v0.26.0 a v0.48.0**. No es un módulo nuevo —ya era indirecto del repositorio, por
`github.com/rogpeppe/go-internal v1.16.0`, que pide v0.26.0— sino una versión mayor, y la pide
`modernc.org/libc v1.75.6`, que el `go.mod` del driver declara:

```
modernc.org/libc@v1.75.6 golang.org/x/tools@v0.48.0
modernc.org/libc@v1.75.6 golang.org/x/sys@v0.47.0
```

Queda, por tanto, dentro de lo que el supuesto permitía —«los indirectos que `go mod tidy` escriba», y
nada más allá de lo que declara el `go.mod` del driver— y la razón por la que la sonda no lo vio es la
que S1 ya daba: el módulo de la sonda no tenía las demás dependencias de `kitlegal`, así que allí no
había ningún `x/tools` con el que hacer selección mínima.

`govulncheck` no encuentra ninguna vulnerabilidad con el árbol nuevo, y `go mod verify` da por buenos
todos los módulos.
