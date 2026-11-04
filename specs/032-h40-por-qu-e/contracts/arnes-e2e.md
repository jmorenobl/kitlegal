# Contrato: el arnés e2e y el guion congelado de la tabla

FR-020, FR-021, FR-023, FR-030; SC-005, SC-007. Research D13; V13 a V17.

## 1. Lo que gana el arnés e2e

`TestMain` de `internal/app/e2e_test.go` construye, además de las siete del binario de e2e de `kitlegal`, el programa
del repositorio `cmd/laboratorio` (`go install` del paquete `github.com/jmorenobl/kitlegal/cmd/laboratorio`, con
`GOBIN` en una carpeta propia del temporal del arnés, como las demás, y con su nombre, `laboratorio`), y cada guion recibe su ruta absoluta en **`KITLEGAL_LABORATORIO_BIN`**. No va en el `PATH`, no cambia ninguna
otra variable, ni el binario de e2e, ni las órdenes del arnés (`cronometra`, `arbol`, `mcp`), ni ningún guion de hoy.
Si no se puede construir, el error queda en `entorno.err` y lo cuenta cada test que necesita los binarios, como hoy.
`TestVariablesDeLosGuiones` la cuenta entre las rutas absolutas que fija; `TestBinariosDelArnes`, que comprueba la
versión, el enlazador y el reloj de las siete construcciones de `kitlegal`, no cambia.

Lo hace la tarea `[aceptacion]`, con el guion: es código de test, no del producto, y sin él el guion no podría fallar
hoy por una aserción (research D13).

## 2. El guion `aceptacion/laboratorio-tabla.txtar`

Activado como `internal/app/testdata/script/h40-laboratorio-tabla.txtar` y ejecutado por `TestEntregaDelHito` en
`make ci`. Un comentario de cabecera nombra lo que cubre —FR-020, FR-021, FR-023, FR-030, SC-005 y SC-007— y no contiene
los textos `usage: `, `unknown command`, `unexpected command` ni `cannot parse`. Órdenes, en este orden:

```text
exec $KITLEGAL_LABORATORIO_BIN tabla h39
cmp stdout h39-tabla.md
! stderr .
exec $KITLEGAL_LABORATORIO_BIN tabla normativa
cmp stdout normativa-tabla.md
! stderr .
```

Ficheros del archivo, en el directorio de trabajo del guion, que es donde `laboratorio tabla` lee la declaración de H39:

- **`laboratorio/variantes.yaml`**: el de H39, copiado de `main`, byte a byte.
- **`h39/h39-a/…`** y **`h39/h39-b/…`**: los 63 ficheros de `internal/evals/testdata/laboratorio/sondeo/` de `main`, con
  sus rutas relativas, byte a byte (también los vacíos).
- **`h39-tabla.md`**: `internal/evals/testdata/laboratorio/tabla.md` de `main`, byte a byte. Es lo que FR-023 fija: si el
  hito cambiara esas sesiones, esa tabla o los tests de H39, este guion, congelado, seguiría comparando con las de `main`.
- **`normativa/h40-a/sondeo.json`**: `{"modelo":"claude-sonnet-5-5","claude_code":"2.1.317"}` y su salto de línea.
- **cinco sesiones de `normativa`**, cada una con `codigo-de-la-sesion`, `sesion.err` y `sesion.jsonl` con la forma de las
  de H39 (`system/init` con el modelo y la versión; un `assistant` con su `message.id` por respuesta; los resultados en
  mensajes `user`; y `result`):

  | Sesión | Lo que hace |
  |---|---|
  | `normativa-control/orden/04-lgt-prescripcion-01` | `Skill` `normativa` (turno 1); `Bash` `cat …/normativa/references/normas.md` (2); `Bash` `kitlegal boe articulo BOE-A-2003-23186 a66 --json` (3); respuesta sin el sello |
  | `normativa-control/herramienta/05-trlrhl-impuestos-municipales-01` | `Skill` `normativa` y `mcp__kitlegal__boe_articulo` en la misma respuesta (1); respuesta sin el sello |
  | `normativa-cuerpo-corto/herramienta/15-irpf-rendimientos-por-materia-01` | sin terminar: el tope de 240 s agotado, código 124, como la de H39 `h39-a/control/herramienta/04-et-vacaciones-01` |
  | `normativa-testigo-en-referencia/herramienta/04-lgt-prescripcion-01` | `Skill` `normativa` (1); `Read` de `…/normativa/references/con_herramientas.md` por tramos (2 y 3); `mcp__kitlegal__boe_articulo` (4); respuesta que termina en `Sello de lectura: KL-7Q4` |
  | `normativa-testigo-en-referencia/herramienta/05-trlrhl-impuestos-municipales-01` | `Skill` `normativa` (1); `mcp__kitlegal__boe_articulo` (2); respuesta que termina en `**Sello de lectura: KL-7Q4**` |

- **`normativa-tabla.md`**: la tabla y el detalle de esas cinco, los de contracts/tabla.md §3 y §4, byte a byte.

Cubre: una referencia abierta dos veces y otra que ninguna sesión abre, `Skill` y la primera herramienta en el mismo
turno, una sesión sin terminar, una fila sin ninguna sesión terminada, y una sesión que cumple la regla testigo y otra
que no (FR-020, FR-021, FR-030); las columnas de la familia y la celda vacía de la regla fuera de
`normativa-testigo-en-referencia`; y la tabla de H39 de `main` (FR-023, SC-007).

**Hoy falla por una aserción**: con el arnés del §1, la parte de H39 pasa, y `laboratorio tabla normativa` sale con 1
(`… no es ninguna variante de la declaración del laboratorio`), que `exec` da por fallo.

## 3. Lo que no cubre

Los extremos de Wilson de 0, 20 y 40 de 40 y el resto de FR-022 los comprueba `TestTablaDeNormativa` (contracts/tabla.md
§7): un sondeo de 40 sesiones por fila no cabe con sentido en un archivo. Las historias 1 y 2 no tienen salida de un
programa que un guion pueda leer sin modelo: su aceptación son `TestVariantesDeNormativa`,
`TestLaboratorioFueraDelProducto` y `TestArnesDeNormativa`, en `make ci`. La historia 4 la hace la persona.
