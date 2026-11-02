# 0049 · Un bloque que el índice guardado de su norma no tiene no existe

- **Estado**: propuesta (2026-10-30); queda aceptada cuando una persona fusiona la propuesta de cambio que la trae.
- **Fecha**: 2026-10-30
- **Hito**: ninguno. Se escribe en una sesión interactiva tras la medición de cierre de H32 (#214), fuera de su run,
  porque el applet `boe` está fuera de su alcance. Cambia qué clase lleva un fallo de `kitlegal boe articulo` y
  `articulos` y la ayuda de sus argumentos de bloque. No cambia el contrato `Applet` (ADR 0005), ni el sobre (ADR
  0006), ni la tabla de resultados (ADR 0023), ni lo que sirve la caché de una sesión de evals (ADR 0043), ni ningún
  umbral (ADR 0038).

## Contexto y problema

El id de un bloque del BOE no siempre es el número del artículo: en la Ley 9/2017 el «Artículo 118» es el bloque
`a1-30`, y `a118` no existe. Las skills lo dicen desde la #207 («copia el id de la entrada cuyo `titulo` es el
artículo; nunca lo compongas con su número»), y aun así la sesión lo compone a veces con el número de la pregunta.

- **En el uso real** el BOE responde 404 al bloque compuesto, `kitlegal` sale con 3 («la norma … no tiene el bloque
  `a118`») y la skill vuelve al índice.
- **En una sesión de evals** el bloque compuesto no está servido, sale con 5 (ADR 0043: la sesión no alcanza el
  `robots.txt` del BOE) y la skill lo trata como la fuente caída: «⚠ SIN CONSULTA AL BOE». La eval 22 de `normativa`
  (el art. 118 de la LCSP el 1 de junio de 2019) lo midió en las tres sesiones de Sonnet 5.5 en modo herramienta de la
  medición de cierre de H32 (#214, ejecución 41024981954): 0 de 3, y `redaccion_sin_fechar` 3 de 6. En la #204 y la
  #207 había sido una de tres.
- **La ayuda del argumento `bloque`**, «Identificador del bloque en el índice de la norma: a21, da-3…», es la de
  `kitlegal boe articulo --help`, y su primer ejemplo es un número de artículo. El modo herramienta no la ve:
  `--describe`, del que salen los esquemas de las herramientas, no lleva la ayuda de los argumentos.

En los dos casos el binario ya sabe, sin pedir nada, que el id no existe, si tiene guardado el índice de la norma: en
las sesiones de evals, el de cada norma de `data/normas.yaml` está servido (ADR 0043, punto 1); en el uso real, lo
está si la sesión lo leyó en los últimos siete días.

## Opciones consideradas

1. **Solo el texto de las skills.** Insuficiente: ya lo decían desde la #207 y la sesión lo compone igual en un modo.
2. **Comprobar el id contra el índice antes de pedir el bloque, y pedir el índice si no está guardado.** Rechazada:
   una petición más en cada lectura sin índice guardado, y un índice guardado de hace días rechazaría un artículo que
   la norma acaba de ganar.
3. **Cuando la lectura del bloque falla, mirar el índice guardado y, si no tiene ese id, salir con 3 y nombrar el
   bloque del artículo.** Ninguna petición más ni ninguna lectura que el BOE sirva rechazada; el fallo dice lo que de
   verdad pasa —el bloque no existe— y lo que hay que pedir.

## Decisión

Opción 3.

1. Si la lectura de un bloque con `articulo` (con `--fecha` o sin ella) o con `articulos` falla con «no encontrado»,
   «fuente no disponible» o «límite o TOS» —también la entrada ausente con `--offline`— y la caché tiene vigente el
   índice de su norma sin ese id, el fallo es «no encontrado», código 3, con la dirección del bloque y el instante del
   fallo. El índice no se pide: sin él guardado, o con el id en él, el fallo es el de hoy. Un fallo inesperado o de
   los argumentos, o con el contexto terminado, tampoco cambia.
2. Si el id tiene la forma de un número de artículo (`a`, el número y, si lo hay, un sufijo en minúsculas: `a118`,
   `a85bis`) y el índice tiene entradas cuyo título es ese artículo —«Artículo 118», «Artículo 85 bis», seguidos o no
   de un punto y su rúbrica—, el mensaje las nombra: «la norma BOE-A-2017-12902 no tiene el bloque a118 en su índice:
   el «Artículo 118» es el bloque a1-30». Si no, remite al campo `id` de `kitlegal boe indice`.
3. La ayuda de `bloque` y `bloques` en `--help` dice que el id sale del campo `id` de `kitlegal boe indice` y no se
   compone con el número del artículo. Los esquemas de las herramientas no cambian.

## Consecuencias

- En las evals, el `5` de un id compuesto con una norma de `data/` pasa a ser el `3` del uso real, con el id bueno en el
  mensaje: la eval mide la respuesta a la avería que se da de verdad, no la que causa su propia caché.
- En el uso real, el 404 del bloque compuesto nombra el bloque bueno si el índice está guardado; la skill que vuelve al
  índice con un 3 lo sigue haciendo.
- En el informe del job, esa lectura deja de contar en `fuera_de_lo_grabado`, que reúne las de código 4 o 5 y solo
  informa, y pasa a `otras_fallidas`: el bloque no falta de lo grabado, no existe. Las `lecturas` de las evals son
  bloques del índice y no cambian. La reconstrucción de los casos del juez, que repite las órdenes de la sesión con
  `--offline` sobre su caché, da el mismo fallo que la sesión.
- Si el BOE añade un artículo y el índice guardado es anterior, la lectura del bloque nuevo no cambia: solo se mira el
  índice cuando la lectura ya ha fallado.
- Lo prueban `TestBloqueFueraDelIndice` y `TestBloquesDelArticulo` (`internal/source/boe/fuera_del_indice_test.go`).
