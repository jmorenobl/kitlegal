# Contrato: el guion de aceptación congelado

Research D15. Lo escribe la primera tarea en `specs/036-h41-2-por-qu-e/aceptacion/` y queda congelado; el workflow lo
activa como `internal/app/testdata/script/h41-2-laboratorio-tabla.txtar` y lo ejecuta `TestEntregaDelHito` en `make ci`.
El arnés e2e no cambia: `KITLEGAL_LABORATORIO_BIN` es el `laboratorio` del árbol (H40). El guion abre con un comentario
que nombra lo que cubre y no contiene los textos `usage: `, `unknown command`, `unexpected command` ni `cannot parse`.
Falla hoy por una aserción: el guion de hoy ignora `skills_instaladas` (research V11) y da la tabla sin las columnas
nuevas, que no es la esperada.

## `aceptacion/laboratorio-tabla.txtar` — historia 3 (FR-020, FR-021, FR-022, FR-024; SC-005)

Órdenes, en este orden; el `sh -c … || true` va delante de la orden con código, para que antes de implementar el guion
falle en la comparación y no por el código (el patrón de H40 y H41):

```text
exec sh -c '"$KITLEGAL_LABORATORIO_BIN" tabla activaciones > activaciones.out || true'
cmp activaciones.out activaciones-tabla.md
exec $KITLEGAL_LABORATORIO_BIN tabla activaciones
cmp stdout activaciones-tabla.md
! stderr .
```

Ficheros del archivo: bajo `activaciones/`, los tres sondeos de contracts/tabla.md §6 —`h41`, `h41-2-a` y `h41-2-b`, con
sus `sondeo.json` y sus ocho sesiones—, y `activaciones-tabla.md`, la tabla de contracts/tabla.md §4, byte a byte. Cada
sesión lleva `codigo-de-la-sesion`, `sesion.err` y `sesion.jsonl` con la forma de las de H41: `system/init` con el modelo y
la versión de su sondeo; un `assistant` con su `message.id` por respuesta, con sus bloques `tool_use`; los resultados en
mensajes `user`; y `result` con la respuesta. Las dos sin terminar, código `124` y sin `result`.

Las historias 1 y 2 no tienen una orden del binario que un guion pueda lanzar: su aceptación en `make ci` es
`TestInformeConLasActivaciones` (contracts/informe.md §5) y `TestArnesConLaInstalacionDelProducto` y
`TestArnesConUnaEval` (contracts/arnes.md §5). La 4 la hace la persona (quickstart.md §8). Que una tabla sin constancia
sale como hoy (FR-022, SC-006) lo guardan las suites congeladas de H40 y H41, sin cambiar un fichero suyo.
