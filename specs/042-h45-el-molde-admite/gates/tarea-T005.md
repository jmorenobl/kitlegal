# T005: por qué el intento 1 no quedó en verde

## Lo hecho (queda en el árbol)

- `internal/skills/testdata/vertical-de-ensayo/data/plazos.yaml`, nuevo: `tipos:` con el tipo `ensayo` (cómputo
  `lpac`, `{cantidad: 10, unidad: dias-habiles}`, `da: vencimiento`, precepto `BOE-A-1900-1` `a2` `"2"`
  `fija: duracion` y `vigente_desde: "1900-01-01"`, que `plazos.LeerTipos` exige en cada precepto,
  `internal/core/plazos/tipos.go:750-760`).
- `internal/skills/testdata/vertical-de-ensayo/data/cobertura.yaml`: la fila «procedimientos de ensayo que la ley común
  deja aparte», precepto `{norma: BOE-A-2015-10565, bloque: daprimera, apartado: "DA 1ª.2.c"}` y
  `norma: {identificador: BOE-A-1900-2, nombre: …, aplicable: {bloque: a1, apartado: "1"}}` (`aplicable` lo exige
  `schemas/cobertura.yaml.json`), y la cabecera al día.
- `internal/skills/testdata/vertical-de-ensayo/skills/plazos-y-territorio/references/ensayo.md`: la sección «Recursos y
  plazos» y las marcas vacías de `recursos_y_plazos`.

Sin la fila nueva de la cobertura (quitada a mano un momento y repuesta), `go test ./internal/skills/ -run
TestVerticalDeEnsayo -count=1` sale en verde: el `plazos.yaml` y la región nueva no dan ningún defecto.

## Por qué cae

`make ci`, visto en esta sesión, cae solo en `internal/skills`:

```text
--- FAIL: TestVerticalDeEnsayo (0.43s)
        	Error:      	Should be empty, but was [plazos-y-territorio: references/administrativo.md: 101 líneas sin un índice que nombre sus apartados ## antes del primero (más de 100)]
FAIL
FAIL	github.com/jmorenobl/kitlegal/internal/skills	8.542s
make: *** [Makefile:97: test] Error 1
```

Hoy una fila de `regulados_aparte` va a la región de la norma de su precepto (`filasDeLaCobertura`,
`internal/skills/regiones.go:511-515`; y `filasSinRegion`, `regiones.go:888-892`, con la misma regla). La fila nueva
tiene su precepto en la LPAC, que es de lo administrativo, así que en la copia de `TestVerticalDeEnsayo` sale en
`skills/plazos-y-territorio/references/administrativo.md`, que en el repositorio tiene 100 líneas y pasa a 101 sin
índice (la regla de las referencias de más de 100 líneas). La línea de la tarea daba por hecho que con el molde de
T003 todo seguía en verde, y no contaba con esa fila: lo que la saca de `administrativo.md` es `paqueteDeLaFila`
(contracts/sincronia-de-los-paquetes.md §3), que estaba en T006, una tarea sin `[datos]` que no puede añadir la fila.

## La tarea, mal delimitada: lo que se amplía

La línea de T005 en `tasks.md` declara ahora también `internal/skills/regiones.go` (y con él su
`internal/skills/regiones_test.go`): el intento siguiente añade `paqueteDeLaFila(entrada, normas)` —el `vertical` de
la `norma` de la fila si está en la tabla de normas con `vertical`; si no, el de la norma de su precepto— y lo usan
`filasDeLaCobertura` y `filasSinRegion` (defecto con el texto de hoy si su paquete es vacío o no está entre los
cubiertos); la línea de cada fila no cambia. El test primero: `TestVerticalDeEnsayo` ya está en rojo con estos datos
y por esta causa, y un subtest de `TestRegiones` en `regiones_test.go` puede fijar la regla antes del código. Con los
datos de `main` las nueve filas siguen siendo de lo administrativo (data-model §3): `make skills-check` sin
diferencias. T006 encuentra hecha esa parte y le queda la de `vertical_de_ensayo_test.go`; su línea no se toca.

## Intento 2: en verde

- Test primero, en `internal/skills/regiones_test.go`: el caso `fila-de-la-norma-de-un-paquete-sin-region` de
  `TestRegiones/defectos` y el subtest `TestRegiones/paquete-de-la-fila`, vistos en rojo por aserción con la regla
  anterior, junto con `TestVerticalDeEnsayo` (101 líneas sin índice).
- `internal/skills/regiones.go`: `paqueteDeLaFila(entrada, normas)`, que usan `filasDeLaCobertura` y
  `filasSinRegion`; la línea de cada fila sigue saliendo de la norma de su precepto.
- Visto en esta sesión: `go test ./internal/skills/ -count=1` (610 en verde), `make skills-check` en verde y
  `make ci` con «ci: todos los controles en verde». `TestVerticalDeEnsayo` pasa sin tocar
  `vertical_de_ensayo_test.go`.
