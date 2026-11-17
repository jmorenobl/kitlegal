# Contrato: el guion de aceptación congelado

Research D12. Lo escribe la primera tarea en `specs/037-h41-3-por-que-normativa/aceptacion/` y queda congelado; el
workflow lo activa como `internal/app/testdata/script/h41-3-laboratorio-tabla.txtar` y lo ejecuta `TestEntregaDelHito` en
`make ci`, como el subtest `TestEntregaDelHito/h41-3-laboratorio-tabla` (research V4). El arnés e2e no cambia:
`KITLEGAL_LABORATORIO_BIN` es el `laboratorio` del árbol (H40). El guion abre con un comentario que nombra lo que cubre y no
contiene los textos `usage: `, `unknown command`, `unexpected command` ni `cannot parse`. Falla hoy por una aserción: el
guion de hoy ignora `instalacion` y `presupuesto` en `sondeo.json`, junta en una fila las de control, diagnóstico y
remedio —mismas skills instaladas, misma versión— y da la tabla sin las dos columnas, que no es la esperada.

## `aceptacion/laboratorio-tabla.txtar` — historia 3 (FR-020, FR-021, FR-022, FR-024; SC-006)

Órdenes, en este orden; el `sh -c … || true` va delante de la orden con código, para que antes de implementar el guion
falle en la comparación y no por el código (el patrón de H40 a H41.2):

```text
exec sh -c '"$KITLEGAL_LABORATORIO_BIN" tabla instalacion > instalacion.out || true'
cmp instalacion.out instalacion-tabla.md
exec $KITLEGAL_LABORATORIO_BIN tabla instalacion
cmp stdout instalacion-tabla.md
! stderr .
```

Ficheros del archivo: bajo `instalacion/`, los cinco sondeos de contracts/tabla.md §6 —`h41-2`, `h41-3-control`,
`h41-3-diagnostico`, `h41-3-remedio` y `h41-3-sola`, con sus `sondeo.json` y sus once sesiones, de la forma que dice esa
sección—, y `instalacion-tabla.md`, la tabla de contracts/tabla.md §4, byte a byte. El comentario de cabecera lleva la
línea `# Cubre: US3 · FR-020, FR-021, FR-022, FR-024 · SC-006`.

Las historias 1 y 2 no tienen una orden del binario ni del programa del laboratorio que un guion pueda lanzar sin
modelo: la 1 es una comprobación de `make ci`, `TestDescripcionesCortas` (contracts/variante.md §4), y la 2 abre sesiones
con modelo y credencial, así que su aceptación en `make ci` son `TestArnesConLasDescripcionesCortas` y
`TestArnesConElPresupuesto`, con el sustituto de `claude` (contracts/arnes.md §6). La 4 la hace la persona
(quickstart.md §7). Que una tabla sin la constancia sale como hoy (FR-022, SC-007) lo guardan las suites congeladas de
H40, H41 y H41.2, sin cambiar un fichero suyo.
