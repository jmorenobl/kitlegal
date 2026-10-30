# Con herramientas

Cómo se pide cada operación del protocolo cuando entre las herramientas del agente están las de kitlegal: cómo se
llaman, cómo van sus argumentos, cuál corresponde a cada orden de `SKILL.md` y qué se hace cuando una falla.

## Nombres

Cada herramienta se llama como la columna «Herramienta» de su fila en «Comandos» de `SKILL.md` —`boe_articulo`,
`graph_check`…—, sola o detrás del prefijo que le ponga el agente, como `mcp__kitlegal__boe_articulo`. Si la tienes,
úsala siempre en lugar de su orden: devuelve lo mismo.

## Argumentos

Van por su nombre, no por su posición: `norma`, el identificador `BOE-A-…`; `bloque`, un id copiado de lo que
devolvió `boe_indice` en esta conversación, nunca compuesto con el número del artículo ni tomado de la pregunta;
`bloques`, varios ids, de ahí mismo; `texto`, las palabras de una búsqueda; y `fecha`, el día `AAAA-MM-DD` de una
lectura con fecha.

## Las órdenes del protocolo

- El índice de una norma: `boe_indice`, con `norma`. Va antes de `boe_articulo` o `boe_articulos` siempre que el id
  del bloque no salga de un `boe_indice` de esta conversación, también con `fecha`.
- Un bloque sin fecha: `boe_articulo`, con `norma` y `bloque`, y detrás `graph_check`, con esa `norma` y ese bloque en
  `bloques`.
- Un bloque en una fecha: `boe_articulo`, con `norma`, `bloque` y `fecha`, sin `graph_check`.
- Varios bloques a la vez: `boe_articulos`, con `norma` y `bloques`, y detrás `graph_check`, con esa `norma` y esos
  mismos bloques en `bloques`.
- La búsqueda de una norma por su título: `boe_buscar`, con las palabras en `texto`.
- Los metadatos y el análisis de una norma: `boe_metadatos` y `boe_analisis`, cada una con `norma`.

## Encadenar y fallar

Donde `SKILL.md` une dos órdenes con `&&`, con herramientas son dos llamadas seguidas, y la segunda solo si la primera
no falló. Donde dice que una orden termina con un código, la herramienta da un resultado marcado como error, y su
`data.clase` dice cuál: `argumentos` para el `2`, `no-encontrado` para el `3`, `fuente-no-disponible` para el `4`,
`limite-o-tos` para el `5`, `identidad-humana` para el `6` e `inesperado` para el `1`. Lo que el protocolo hace con un
código, se hace igual con su clase. Un error de `boe_articulo` con un `bloque` que no salió de `boe_indice` se
resuelve con `boe_indice`, sea cual sea su `data.clase`: lo primero que puede fallar es el id.
