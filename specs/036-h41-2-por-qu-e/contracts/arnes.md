# Contrato: el arnés con la instalación del producto y una sola eval

FR-010 a FR-015; SC-003, SC-004, SC-009. Research D7 a D11; data-model §3. Lo que no se nombra aquí es el arnés de H41 tal
cual (`specs/034-h41-las-skills-del/contracts/arnes.md`).

## 1. Cómo se lanza

La misma orden y la misma cadena (`Makefile` → `scripts/evals-sondeo-llavero.sh laboratorio` →
`scripts/evals-laboratorio.sh` → `go test -tags evals … -run '^TestSondeoDelLaboratorio$' ./internal/evals/` →
`sondearElLaboratorio`), con dos argumentos más, detrás de los siete de hoy:

```sh
make evals-laboratorio VARIANTES=<v>[,<v>…] MODOS=<m>[,<m>] SESIONES=<n> MODELO=<id> CLAUDE_CODE=<versión> \
  CONCURRENCIA=<c> DIRECTORIO=<ruta> [INSTALACION=producto] [EVAL=<nn>]
```

La receta pasa `"$(INSTALACION)" "$(EVAL)"` detrás de `"$(DIRECTORIO)"`; `scripts/evals-laboratorio.sh` exige nueve
argumentos y los da a `go test` como `-instalacion` y `-eval`. Lo lanza una persona con la credencial del sondeo; nunca
`make ci`, un gancho, el job ni un paso del workflow (FR-014).

Las cuatro filas de FR-030 son cuatro invocaciones y una tabla; `CLAUDE_CODE` es la versión que la persona tiene en el
`PATH` en cada una, y el arnés comprueba que lo es antes de abrir nada:

```sh
make evals-laboratorio VARIANTES=normativa-producto MODOS=herramienta SESIONES=40 MODELO=claude-sonnet-5-5 \
  CLAUDE_CODE=2.1.284 CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h41-2/sola-2.1.284 EVAL=04
make evals-laboratorio VARIANTES=normativa-producto MODOS=herramienta SESIONES=40 MODELO=claude-sonnet-5-5 \
  CLAUDE_CODE=2.1.284 CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h41-2/todas-2.1.284 EVAL=04 INSTALACION=producto
make evals-laboratorio VARIANTES=normativa-producto MODOS=herramienta SESIONES=40 MODELO=claude-sonnet-5-5 \
  CLAUDE_CODE=<la del día> CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h41-2/sola-<la del día> EVAL=04
make evals-laboratorio VARIANTES=normativa-producto MODOS=herramienta SESIONES=40 MODELO=claude-sonnet-5-5 \
  CLAUDE_CODE=<la del día> CONCURRENCIA=5 DIRECTORIO=../kitlegal-sondeos/h41-2/todas-<la del día> EVAL=04 INSTALACION=producto
make laboratorio-tabla DIRECTORIO=../kitlegal-sondeos/h41-2
```

4 × 40 = 160 sesiones, todas `04-lgt-prescripcion-01` a `-40`.

## 2. Argumentos

Todos los errores van juntos, uno por línea, cada uno con el nombre del argumento de `make`; con cualquiera no se abre
ninguna sesión, no se construye nada y no se crea `DIRECTORIO` (FR-011). Los de hoy no cambian ni de texto ni de orden;
los dos nuevos van detrás del de `DIRECTORIO`. Vacíos, los dos son «sin la opción» y «sin el argumento», y no dan error.

| Argumento | Vale si | Mensaje si no |
|---|---|---|
| `INSTALACION` | vacío, o `producto` | `INSTALACION: «<v>» no es producto` |
| `INSTALACION=producto` | las variantes pedidas son de una familia del producto | `INSTALACION: las variantes del laboratorio de H39 no son de ninguna familia del producto` |
| `EVAL` | vacío, o el número de dos cifras de una de las cuatro evals del plan de la familia pedida: la que empieza por `<nn>-` (`04`, `05`, `15` o `22` en `normativa`; `02`, `05`, `10` u `11` en `plazos-y-territorio`; `01`, `03`, `04` o `06` en `jurisprudencia`) | `EVAL: «<v>» no es el número de ninguna eval del plan de <skill>` |
| `EVAL` no vacío | las variantes pedidas son de una familia del producto | `EVAL: las variantes del laboratorio de H39 no tienen el plan de una familia del producto` |
| `SESIONES`, con `EVAL` | un entero mayor que 0 | `SESIONES: «<n>» no es un entero mayor que 0` |
| `SESIONES`, sin `EVAL` | como hoy: un múltiplo de 4 mayor que 0 | como hoy |

Tres de esos errores dependen de la familia de las variantes pedidas: el de `INSTALACION=producto` y el de `EVAL` no vacío
con variantes de H39, y el de un `EVAL` que no es el número de ninguna eval del plan de la familia. Solo se dan cuando las
variantes pedidas que valen son de una sola familia (`familiaDeLasVariantes`, `laboratorio_sondeo.go:470-487`); si ninguna
vale, o las que valen mezclan familias, no hay familia, y el error de `VARIANTES` o el de la mezcla ya lo dicen. El de
`INSTALACION: «<v>» no es producto` no depende de la familia: sale siempre que el valor no es vacío ni `producto`, con
cualquier `VARIANTES`. La regla de `SESIONES` la decide que `EVAL` venga o no, valga o no. Dos ejemplos, en la salida de
error. El de `VARIANTES=control MODOS=orden SESIONES=0 MODELO=claude-sonnet-5-5 CLAUDE_CODE=2.1.284 CONCURRENCIA=1
DIRECTORIO=<uno que no existe> INSTALACION=todas EVAL=04`:

```text
SESIONES: «0» no es un entero mayor que 0
INSTALACION: «todas» no es producto
EVAL: las variantes del laboratorio de H39 no tienen el plan de una familia del producto
```

Y el de `VARIANTES=otra MODOS=herramienta SESIONES=4 MODELO=claude-sonnet-5-5 CLAUDE_CODE=2.1.284 CONCURRENCIA=1
DIRECTORIO=<uno que no existe> INSTALACION=todas`, sin familia:

```text
VARIANTES: «otra» no es ninguna variante del laboratorio
INSTALACION: «todas» no es producto
```

## 3. Lo que prepara y abre

Lo de H41, con tres diferencias:

1. **Skills** (FR-010). En el temporal, `skills/<variante>/` tiene la copia de la variante, como hoy. Con
   `INSTALACION=producto`, tiene además la de cada directorio de `skills/` del árbol que no es la skill de la familia, con
   `os.CopyFS`; en `<skill>-producto`, eso es `skills/` entera. Cada sesión enlaza en `claude/skills/` cada entrada, así que
   ve las cuatro skills del producto —la de su familia, la de su variante—, cada una byte a byte, sin ningún nombre
   repetido. Sin la opción, la de su variante sola, byte a byte como hoy.
2. **Plan** (FR-013). Con `EVAL`, el plan de cada par variante × modo son `SESIONES` sesiones de esa eval, de nombre
   `<eval sin .yaml>-<nn>`, desde `01`. Sin él, `SESIONES/4` de cada una de las cuatro evals del plan, como hoy.
3. **`sondeo.json`** (FR-012). En una familia del producto, detrás de lo de hoy, `skills_instaladas`: los nombres de las
   entradas de `skills/<variante>/` del temporal, en orden de nombre (data-model §3). En H39, el de hoy. Ni la copia ni su
   huella se guardan en `DIRECTORIO`.

La versión de Claude Code y el modelo se comprueban y se anotan como hoy (FR-014). Ejemplo, el `sondeo.json` de la segunda
invocación de §1:

```json
{"modelo":"claude-sonnet-5-5","claude_code":"2.1.284","referencias_del_producto":["buscar_la_norma.md","como_se_cita.md","con_herramientas.md","lecturas_que_fallan.md","normas.md","redaccion_en_una_fecha.md","redaccion_modificada.md","varios_bloques.md"],"skills_instaladas":["consulta-juridica","jurisprudencia","normativa","plazos-y-territorio"]}
```

## 4. Salida y uso

La salida del arnés no cambia de forma: la de la segunda invocación de §1 es

```text
Sondeo del laboratorio en /Users/…/kitlegal-sondeos/h41-2/todas-2.1.284, con claude-sonnet-5-5 y Claude Code 2.1.284.
- normativa-producto, herramienta: 40 abiertas.
La tabla: make laboratorio-tabla DIRECTORIO=/Users/…/kitlegal-sondeos/h41-2/todas-2.1.284
```

La lee la persona al terminar cada invocación: tres líneas en las de FR-030. Con un error de uso, la salida de error es la
lista de §2, de 40 a 130 bytes por argumento que no vale, que deja de darse cuando la orden vale. `sondeo.json`, de 89 a
349 bytes (research V16), lo lee el guion de la tabla una vez por sondeo. No crecen con el uso del producto —cientos de
normas y miles de bloques consultados no los cambian—, solo con las variantes, modos y skills pedidos. Ninguna skill del
producto los pide.

## 5. Controles

- **`TestArnesConLaInstalacionDelProducto`** (`internal/evals/laboratorio_sondeo_test.go`, en `make ci`, sin modelo;
  FR-010 a FR-012, FR-015; SC-003, SC-004): `sondearElLaboratorio` con la construcción sustituida y el sustituto de
  `claude` de `TestArnesDelLaboratorio`, que anota sus invocaciones.
  - Con `INSTALACION=producto`, una invocación por familia con su `<skill>-control` y su `<skill>-producto`
    (`MODOS=herramienta SESIONES=4`): en cada sesión, `claude/skills/` tiene exactamente `consulta-juridica`,
    `jurisprudencia`, `normativa` y `plazos-y-territorio`; la de la familia es, ruta a ruta y byte a byte, la de
    `laboratorio/<skill>/<skill>-control/skills/<skill>` o la de `skills/<skill>`, y cada otra, la de `skills/<otra>`; y
    `sondeo.json` lleva las cuatro en `skills_instaladas`. Falla con la primera sesión cuyas skills difieren.
  - Con `INSTALACION=todas`, con `INSTALACION=todas` y `VARIANTES=otra` (sin familia), y con `INSTALACION=producto` y
    `VARIANTES=control` (H39): 0 sesiones, nada construido, `DIRECTORIO` sin crear y sus mensajes de §2; con
    `VARIANTES=otra`, los dos del segundo ejemplo, en ese orden.
- **`TestArnesConUnaEval`** (`laboratorio_sondeo_test.go`; FR-011, FR-013, FR-015; SC-004, SC-009): con `EVAL`, una
  variante de cada familia —`normativa-producto` con `EVAL=04 SESIONES=3`, `plazos-y-territorio-control` con
  `EVAL=05 SESIONES=2` y `jurisprudencia-producto` con `EVAL=03 SESIONES=1`, en modo herramienta— abre solo
  `<eval>-01` a `<eval>-<SESIONES>` de esa eval, y ninguna más; y las dos invocaciones de FR-030 con `SESIONES=2`, con
  `INSTALACION=producto` y sin ella. Con `EVAL=07` en `normativa`, `EVAL=4`, `EVAL=04` y `VARIANTES=control`, y `EVAL=04`
  con `SESIONES=0`: 0 sesiones y su mensaje de §2.
- **`TestArnesDeLasFamilias`**, **`TestArnesDeNormativa`** y **`TestArnesDelLaboratorio`** (H39 a H41), sin la opción ni
  el argumento: las mismas sesiones y, en cada una, las mismas skills, byte a byte, sin cambiar ningún valor esperado de
  lo que instalan ni de lo que abren. Cambia solo el `sondeo.json` que exige `exigirElSondeoDelArnes` en las familias del
  producto, que gana `"skills_instaladas":["<skill>"]` (research D11); el de H39, no.
- **`TestGuionDelLaboratorio`** (`laboratorio_sondeo_test.go`, en `make ci`, con el sustituto de go): sus casos pasan
  nueve argumentos a `scripts/evals-laboratorio.sh`, uno con `producto` de octavo y otro con `04` de noveno y los demás
  con los dos vacíos; `go` recibe `-directorio <directorio> -instalacion <octavo> -eval <noveno> -temporal <temporal>`,
  con `-temporal` el último; con ocho, el guion sale con 1, da su línea de uso, que nombra `<instalacion> <eval>` detrás
  de `<directorio>`, y no ejecuta go.
- **`TestRecetasDelSondeo`**: su caso de `evals-laboratorio` pasa también `INSTALACION=producto EVAL=04` y espera la
  orden del llavero con los nueve argumentos.
- **`TestSondeoDelLaboratorio`** (`-tags evals`): la entrada de la persona, con `-instalacion` y `-eval`; no corre en
  `make ci`.
