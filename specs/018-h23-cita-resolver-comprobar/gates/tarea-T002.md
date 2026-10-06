# T002: por qué no quedó en verde

## Intento 1: tarea mal delimitada, redelimitada (2026-10-06)

**Causa.** `TestSuperficieDeIds/lo_exportado_es_el_contrato`, de H6, vive en
`internal/core/ids/errores_test.go` y fija la lista cerrada de lo que exporta el paquete
(`superficieDelContrato`). T002 manda exportar `ECLI`, `AnalizarECLI`, `ECLI.String`, `ECLI.Organo`, `ROJ`,
`AnalizarROJ` y `ROJ.String`, así que ese test falla por construcción, y su fichero no estaba entre las rutas de
la tarea (`ecli.go`, `roj.go`, `doc.go` y sus `_test.go`): `errores.go` no se declara y `errores_test.go` no es
el test de ningún fichero declarado. No hay forma de cumplir la tarea sin cambiar esa lista. Ni el plan ni
research D30 («Las listas literales») la nombran.

**Qué se ha hecho en este intento, dentro de las rutas declaradas, y queda en el árbol:**

- `internal/core/ids/ecli_test.go` y `roj_test.go`, escritos primero (rojo: no compilaban), con
  `TestAnalizarECLI`, `TestAnalizarROJ`, `FuzzECLI` y `FuzzROJ`; las semillas van con `f.Add` y la forma de
  FR-003 y de FR-004 está escrita en el test como expresión regular, aparte del analizador, que comprueba byte a
  byte.
- `internal/core/ids/ecli.go`, `roj.go` y `doc.go`.
- `.golangci.yml`: una entrada más en la lista de palabras españolas de `misspell`, `constitucional` («Tribunal
  Constitucional» lo leía como «Constitutional»). Es el único cambio del fichero; el guardián lo admite sin
  declararlo cuando el diff solo añade entradas a esa lista.
- `errores_test.go` **no se ha tocado**: no era de la tarea en este intento.

**Qué falta, y es todo lo que falta:** aplicar el diff de `errores_test.go`, ya verificado, que está en
`specs/018-h23-cita-resolver-comprobar/gates/tarea-T002-errores_test.diff`:

```sh
git apply specs/018-h23-cita-resolver-comprobar/gates/tarea-T002-errores_test.diff
make ci
```

El diff hace tres cosas en ese fichero: `superficieDelContrato` gana las siete declaraciones nuevas y su
comentario dice de dónde sale cada parte; `TestSuperficieDeIds` deja de nombrar el ECLI entre lo que todavía no
ha entrado, y su caso sintético «tipo que no es estructura» pasa de `type ECLI string` a `type CELEX string`; y
`TestClaseDeLosErrores`, que dice pasar «por cada camino de rechazo del paquete», gana una fila por cada uno de
los doce del ECLI y de los nueve del ROJ. Con `make ci` en verde, se marca la tarea `[X]`.

La línea de T002 en `tasks.md` ya declara `internal/core/ids/errores_test.go` y describe ese cambio. Comprobado
con el filtro de rutas de `scripts/workflow/tarea.sh`: la línea da ahora esa ruta, además de las tres de antes.

**Verificado en este intento** (todo ejecutado en primer plano y visto terminar):

- En el repositorio, `go test -count=1 -race -run '^(TestAnalizarECLI|TestAnalizarROJ|FuzzECLI|FuzzROJ)$' -v
  ./internal/core/ids/`: `ok`, 86 líneas `--- PASS`.
- En una copia desechable del árbol bajo el directorio temporal (`rsync -a`, con `.git`), con el diff de
  `errores_test.go` aplicado: `go test -count=1 -race -cover ./internal/core/ids/` da `ok` con
  `coverage: 100.0% of statements`; `make fmt-check lint`, `0 issues.`; y **`make ci`, «ci: todos los controles
  en verde»**.
- En esa copia, `go test -run '^$' -fuzz '^FuzzECLI$' -fuzztime 30s` (4 345 697 ejecuciones) y lo mismo con
  `FuzzROJ` (3 331 758): `PASS` los dos, y `testdata/fuzz/` sigue con sus dos directorios de H6.
- `git apply --check` del diff sobre el árbol del repositorio: se aplica.
- En el repositorio, `make ci` en primer plano: **rojo**, y solo por ese test. `fmt-check` y `lint` pasan
  (`0 issues.`), todos los demás paquetes dan `ok`, y `make` se detiene en `test`:

```
--- FAIL: TestSuperficieDeIds (0.00s)
    --- FAIL: TestSuperficieDeIds/lo_exportado_es_el_contrato (0.01s)
        errores_test.go:180:
            	Error:      	elements differ
            	            	extra elements in list B:
            	            	([]interface {}) (len=7) {
            	            	 (string) (len=41) "type ECLI struct{ /* campos privados */ }",
            	            	 …
            	            	 (string) (len=28) "func (r ROJ) String() string"
            	            	}
            	Messages:   	lo exportado por internal/core/ids no es el contrato de identificadores §1: ni ELI, ni ECLI, ni CELEX, ni NIF, que entran con sus hitos (FR-034)
FAIL
coverage: 100.0% of statements
FAIL	github.com/jmorenobl/kitlegal/internal/core/ids	1.985s
make: *** [test] Error 1
```

**No verificado en el intento 1:** `scripts/workflow/guardian-diff.sh tarea` no se ha podido ejecutar en esa sesión (pide
aprobación). Leído su código, lo cambiado fuera del directorio del feature son `doc.go`, `ecli.go`, `roj.go`,
`ecli_test.go` y `roj_test.go` de `internal/core/ids/` y la entrada de `.golangci.yml`, que son las rutas
declaradas, sus tests y la excepción de la lista de `misspell`; pero eso es una lectura, no una ejecución.
Los objetivos de `make ci` posteriores a `test` (`test-integration`, `test-tiempos`, `vuln`, `schema-check`,
`skills-check`, `goreleaser-check`, `secrets`, `mod-verify`, `mod-tidy-check`) no han corrido en el repositorio,
porque `make` se detuvo antes; en la copia, con el diff aplicado, pasaron todos.

## Intento 2: en verde, marcada `[X]` (2026-10-06)

Lo único que faltaba era el diff de `errores_test.go`. Todo lo del intento 1 seguía en el árbol sin cambios.

**Hecho y verificado en este intento** (todo en primer plano, en el repositorio, y visto terminar):

- Rojo primero: `go test -count=1 -race -run '^TestSuperficieDeIds$' ./internal/core/ids/` falla por el motivo
  documentado arriba (las siete declaraciones del ECLI y del ROJ como «extra elements»).
- `git apply specs/018-h23-cita-resolver-comprobar/gates/tarea-T002-errores_test.diff`: se aplica tal cual.
- `go test -count=1 -race -cover ./internal/core/ids/`: `ok`, `coverage: 100.0% of statements`.
- `TestClaseDeLosErrores` con `-v`: 21 líneas `--- PASS` con prefijo `ECLI` o `ROJ` (las doce del ECLI y las
  nueve del ROJ que añade el diff) y ninguna `--- FAIL`. `FuzzECLI` y `FuzzROJ` sobre sus semillas: `PASS`.
- **`make ci` en el repositorio, salida 0: «ci: todos los controles en verde»**. `fmt-check` y `lint` pasan,
  `test` da `ok` en todos los paquetes (`internal/core/ids`, 100,0 %), y `test-integration`, `test-tiempos`,
  `vuln` (0 vulnerabilidades alcanzadas), `schema-check`, `skills-check`, `goreleaser-check`, `secrets` (sin
  fugas), `mod-verify` y `mod-tidy-check` pasan.

**Fuera de las rutas declaradas** no hay más cambio que el del intento 1 en `.golangci.yml` (una entrada de la
lista de palabras españolas de `misspell`). No se ha añadido ningún supuesto: este intento no abrió ninguna
ambigüedad nueva. `scripts/workflow/guardian-diff.sh tarea` sigue sin ejecutarse en una sesión de un paso: lo
ejecuta el workflow.
