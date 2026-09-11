# Nota de T014 — cómo se contrastaron los guiones, y las dos cosas que no salieron como se escribieron

**Estado de T014: verde.** Esta nota no describe un fallo: describe el contraste que la tarea exige
(«verificado construyendo el binario a mano … y contrastando salida, descriptores y códigos con lo que
afirma cada guion») y deja registradas dos decisiones que conviene conocer antes de T015 y T016.

## El rojo → verde de una tarea que no puede escribir un `_test.go`

T014 es material de test: su ruta es `internal/app/testdata/`, donde los comodines de Go no descienden, y
no puede escribir ningún `_test.go` que `go test ./...` recoja. El control se ejerció en el orden que pide
el ritual:

1. **Rojo.** Los siete guiones `script/*.txtar` se escribieron **primero**, cuando ni el paquete de
   applets ni el binario existían: `go build ./internal/app/testdata/kitlegal-e2e` respondía
   `directory not found`. La entrega estaba descrita y no se podía ni compilar.
2. **Verde.** Escritos `ejemplo/` y `kitlegal-e2e/`, el binario se construyó a mano y cada guion se
   contrastó contra él, aserción por aserción, con un arnés mínimo **desechable** bajo `bin/`
   —que está en `.gitignore`— borrado al terminar. Los siete pasan.
3. **Control negativo.** El mismo arnés, apuntado al binario **distribuido** —que no registra ningún
   applet (FR-009)—, hace fallar **los siete** guiones. Sin él, un guion que no afirmara nada también
   habría «pasado».

El arnés no se versiona a propósito: el ejecutor de verdad es `testscript`, y lo aporta T015. Aquí solo
hacía falta demostrar que lo que cada guion afirma es lo que el binario hace.

## Dos aserciones se escribieron mal y el contraste las corrigió

Son exactamente lo que justifica el paso 2, y quedan anotadas porque las dos son fáciles de repetir:

| Guion | Lo que se escribió | Lo que hace el binario |
|---|---|---|
| `solo-json.txtar` | `stderr 'level=DEBUG'` | El evento de una invocación **correcta** se emite al nivel **informativo**; lo que `KITLEGAL_LOG=debug` añade son los `argumentos`, no el nivel del evento (`internal/cli/log.go`, FR-039) |
| `argumentos.txtar`, `solo-json.txtar` | `"mensaje":"[^"]+"` | El mensaje de un sobre de fallo lleva **comillas escapadas** (`\"noexiste\"`), así que una clase negada de comillas corta el patrón antes de tiempo |

## `Descripcion` y `misspell`: la única excepción de lint que llevan los applets de ejemplo

La tarea pide los dos applets **«sin ninguna excepción de lint»**, y así están respecto de lo que esa
frase significa en este hito ([research.md D19](../research.md), `plan.md` fila 10): `ejemplo/` no recibe
**ninguna** excepción por ruta en `.golangci.yml`, no usa `os.Exit`, `os.Stdout`, `os.Stderr` ni
`fmt.Print*`, y `make lint` lo recorre entero con el mismo listón que el resto del árbol —comprobado: el
paquete aparece nombrado en la orden de `golangci-lint`—.

Llevan, eso sí, **cinco `//nolint:misspell`**, uno por cada aparición del identificador `Descripcion`, que
el contrato del applet fija y el diccionario —solo inglés— lee como errata de `Description`
(`misspell@v0.8.0/words.go:6856`). No hay forma de evitarlo dentro de las rutas de T014: el identificador
lo impone `app.Applet`, y la alternativa buena —añadir `descripcion` a `misspell.ignore-rules`— **no es
media solución sino dos inseparables**, porque en cuanto la regla entra los `//nolint` quedan sin uso y
`nolintlint` los rechaza. Es la deuda que ya describe [`nota-T009-misspell.md`](./nota-T009-misspell.md),
y T014 la **agranda de once a dieciséis**, repartidas ahora en siete ficheros: los cinco de
`internal/app` más `testdata/ejemplo/echo.go` (2) y `testdata/ejemplo/contar.go` (3).

Que la agrande es justo lo que la nota de T009 anticipaba —«cada applet que se escriba a partir de H4
necesitará su propia excepción»—, con la diferencia de que aquí el applet afectado es **la implementación
de referencia que todos van a copiar**. Saldarla sigue necesitando una tarea que declare `.golangci.yml`
**y** esos siete ficheros entre sus rutas; T016 declara el primero pero no los otros seis.

## Lo que este material fija para T015

- El binario del e2e tiene que llamarse **`kitlegal`** en el `PATH` que el test antepone: cuatro guiones
  lo invocan por ese nombre a través del intérprete de órdenes.
- Los guiones usan `exec sh -c '…; test $? -eq 2'` allí donde el **código exacto** es el contrato.
  `testscript` v1.16.0 no tiene ninguna orden que compruebe un código de salida concreto —solo distingue
  éxito de fallo (`testscript/cmd.go:29-52`)—, y renunciar al código exacto habría convertido «termina en
  2» en «termina mal». No hace falta ninguna orden propia en `Params.Cmds`.
- `$KITLEGAL_BIN` se usa tal cual, absoluto, en `exec` y en `symlink`: `symlink` no convierte su destino
  a absoluto (`testscript/cmd.go:462-472`), de modo que el enlace apunta al binario real.
