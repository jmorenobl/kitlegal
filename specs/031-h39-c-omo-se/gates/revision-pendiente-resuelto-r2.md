# Pendiente de la revisión final de H39

## Juez B, motivo 1 ([b][k]), punto (1): sesiones fijas del sondeo

**Lo que pide.** Añadir o reordenar sesiones fijas en `internal/evals/testdata/laboratorio/sondeo/` y su
`internal/evals/testdata/laboratorio/tabla.md` (T009, `[datos]`) para que reconocer la orden cambie «Activa antes»:
una tubería, una secuencia con `&&` y otra con `;`, una orden con `2>&1`, y `--describe`, `--help` y `skills list` en
el mismo turno que la activación o en uno anterior, más una orden `kitlegal` dentro de `$( )` o de una subshell. Y
repetir los tres mutantes (quitar `invocacion.Consulta`, dejar de partir por `|`, dejar de partir por `&` y `&&`)
hasta que fallen.

**Por qué no se ha hecho ahí.** El corrector no puede modificar `testdata/`: el workflow deshace cualquier cambio en
él. Las sesiones fijas y `tabla.md` quedan como están, y `TestTablaDelLaboratorio/sondeos-fijos` sigue dando
`tabla.md` byte a byte con el lector corregido.

**Lo que sí se ha hecho, fuera de `testdata/`.** `TestOrdenesDeBash` (internal/evals/laboratorio_tabla_test.go), con
las órdenes como constantes del test, fija las órdenes simples de `ordenesSimples` y si `esDeKitlegal` cuenta cada
orden en el modo orden, que es lo único que esos mutantes tocan de la decisión «Activa antes». Los tres mutantes
del juez salen en rojo en esta sesión con `go test -overlay`: el de `invocacion.Consulta` por `describe`, `ayuda` y
`dry-run`; el de `|` por `tuberia`, `o` y `redireccion`; el de `&` y `&&` por `y` y `segundo-plano`.

**Lo que queda para una persona.** Si se quiere además que las sesiones fijas de `sondeo/` den un «Activa antes»
distinto según se reconozca o no la orden (la orden `kitlegal` del modo orden en un turno anterior a la activación),
se añaden en una propuesta que pueda tocar `testdata/`, con su fila en `tabla.md`.
