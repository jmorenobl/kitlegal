# Contrato: las raíces de grabaciones y el trabajo de evals de `plazos-y-territorio`

## 1. `UnionDeGrabaciones` (FR-015, §C.4)

```go
// internal/evals/grabaciones.go
func UnionDeGrabaciones() ([]string, error)               // lee ../../evidencias
func unionDeGrabacionesDe(raiz string) ([]string, error)  // la de UnionDeGrabaciones, con la raíz del repositorio dada
func carpetaDeUnHito(nombre string) (hito, subhito int, es bool)
func esDeUnaRaizDeGrabaciones(ruta string) bool           // internal/evals/dependencias.go, sustituye a raicesDeGrabaciones en rutaQueMide
```

- Devuelve, cada llamada en una lista nueva (`dependencias.go:279` añade a ella), `GrabacionesDeH4`, `GrabacionesDeH5`,
  `GrabacionesDeH9` y `raiz + "/evidencias/<carpeta>/boe.legislacion-consolidada"` de cada carpeta de un hito con
  `manifiesto.json` y esa carpeta, por (hito, subhito) en número. Con la raíz `../..`, la de h36 es
  `../../evidencias/h36/boe.legislacion-consolidada`, la constante de hoy, que desaparece.
- `carpetaDeUnHito`: `h36` → (36, 0, sí); `h9-1` → (9, 1, sí); `adr-0036`, `sondeos-20261025` → no. Otros nombres no los
  escribe `grabar-datos.sh` (umbral de materialidad): no se prueban aparte.
- Error: el de `os.ReadDir` de `evidencias/`, envuelto con la ruta. Los llamadores lo propagan:
  `medida.go:1158` y `:1244`, `sesiones.go:441`, `dependencias.go:819` y los tests (también
  `internal/skills/normas_test.go:534`).
- `rutaQueMide` no lee el disco: una ruta es de una raíz si empieza por una de las tres fijas, por `GrafosPrevios` o por
  `evidencias/<carpeta de un hito>/boe.legislacion-consolidada/`.
- Uso: una llamada por preparación de una sesión del job de evals y por test; crece con una raíz por hito que graba el
  BOE (hoy 4), no con el uso.

Tests (SC-005):

| Test | Comprueba | Rojo con |
|---|---|---|
| `TestGrabacionesSinSolape` (`grabaciones_test.go:270`) | sobre el repositorio: empieza por H4, H5 y H9; lleva la de h36; ninguna bajo `evidencias/h9/`; tiene 3 + tantas como carpetas `evidencias/*/boe.legislacion-consolidada` con su `manifiesto.json` al lado (hoy 4); cada dos raíces solo comparten el `robots.txt` (lo de hoy) | un fichero grabado repetido entre dos raíces; `evidencias/h9/` dentro |
| `TestManifiestoDeGrabaciones/raices-de-los-hitos/con-h99` (nuevo, `grabaciones_test.go:75`) | `unionDeGrabacionesDe` sobre un árbol en `t.TempDir()` con `h9` (manifiesto y carpetas de festivos), `h36` y `h99` (manifiesto y carpeta del BOE): 5 raíces, las tres fijas, `h36` y `h99` la última (US7.2) | `h99` fuera o delante de `h36`; `h9` dentro |
| `TestManifiestoDeGrabaciones/raices-de-los-hitos/orden-y-carpetas-que-no-entran` (nuevo) | con `h9-1`, `h36` y `h100` (manifiesto y carpeta del BOE), `h42` (manifiesto sin la carpeta), `h43` (la carpeta sin manifiesto), `adr-0001` y `sondeos-1` (los dos): las tres fijas y detrás `h9-1`, `h36`, `h100` | el orden alfabético (h36 detrás de h100, US7.3); una carpeta sin manifiesto o sin la del BOE dentro; `adr-*` o `sondeos-*` dentro |

## 2. El trabajo `evals` (FR-021, FR-022)

`.github/workflows/evals.yml`, `matrix.include` del trabajo `evals`: `plazos-y-territorio` con `concurrencia: 3`; el
comentario (`:224-242`) da la cuenta con tres y «doce sesiones entre los cuatro», en lugar de «once». Los demás
ajustes, el tope de 352 min y los modelos no cambian.

`internal/evals/definicion_test.go`:

- `sesionesALaVezComoMucho = 12`, con su comentario (1, 3, 5 y 3; sustituye a FR-020 y SC-011 de H30).
- `ajustesDelTrabajo["plazos-y-territorio"] = {Concurrencia: 3, ObjetivoDeDuracion: 0}`; la definición sintética del
  contrato (`:1445`) con la misma concurrencia que el repositorio (`:1673-1676`).
- `probarElPeorCasoDelTrabajo` (`:3194`): para cada skill, lo esperado sale de `LeerConjunto(evals/<skill>)`: por grupo,
  las sesiones de las evals con modo (una serie por modelo que decide y por informativo que no es `Informativa`), por
  repeticiones, más la de la prueba de red en `orden`, y la concurrencia de la definición leída; se compara con lo que da
  `peorCaso` y se exige que quepa en el tope. Sin `[]int{67, 66, 6}`, `20725 * time.Second` ni sus textos.
- Subtest `cuatro-evals-mas`: `evals/` copiada a `t.TempDir()` con cuatro copias de más de una eval de modo de
  `plazos-y-territorio`; con la definición del repositorio, su peor caso cabe en 352 min (cuenta de research D9: 18 821 s
  a tres), y con la misma definición y `plazos-y-territorio` a dos no cabe (27 253 s): el test exige las dos cosas, así
  que la concurrencia de hoy lo pone en rojo dentro de `make ci`.
- `tope-con-las-evals-del-repositorio` (`:1598-1732`): cada caso calcula sus líneas desde el peor caso de la definición
  que monta, no con minutos escritos.
- `internal/evals/plan_test.go:295` (`probarElPlanDelRepositorio`): las series y sesiones por modo, desde
  `LeerConjunto`, como arriba.

Rojo (SC-006): `concurrencia: 2` para `plazos-y-territorio` con las cuatro de más (27 253 s > 21 120 s); 13 sesiones a la
vez (una más en cualquier skill del `matrix.include`).

El informe del trabajo no cambia de forma; doce sesiones a la vez pueden topar con el límite de la suscripción, que el
cierre enseña como fallo de la fuente (ADR 0062, «En contra»).
