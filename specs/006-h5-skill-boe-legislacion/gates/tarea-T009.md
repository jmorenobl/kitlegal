# T009 · intento 1 · en verde

La nota anterior de este intento dejaba la tarea sin marcar porque faltaban las búsquedas grabadas de H5: la grabación
de la pausa de T008 se había cortado con el buscador del BOE caído. Lo resolvió la persona antes de reanudar:

- `1809a56` (T007): el arnés siembra desde H4 con la grabación activa, así que lo que H4 ya grabó no vuelve a la red;
- `1f4feea` (T008): grabación de la pausa, 34 respuestas con estado 200 y ninguna con el nombre de una de H4.

## Vocabulario de `rango`

El `enum` de `schemas/normas.yaml.json` es el conjunto de `rango.texto` de los 70 resultados de las doce búsquedas
grabadas (dos de H4 y diez de H5), leído de las respuestas con `jq` y copiado tal cual: es el mismo texto que da `boe
buscar` (`leerBusqueda` lee `rango.texto` sin transformarlo). Dos búsquedas no aportan resultados (`data` vacío): la
`zzqxkwvjh` de H4 y la de H5 `texto refundido haciendas locales`, que el manifiesto sustituyó en la pausa. Los nueve
valores, en orden de bytes:

`Constitución`, `Decreto Foral Legislativo`, `Ley`, `Ley Foral`, `Orden`, `Real Decreto`,
`Real Decreto Legislativo`, `Real Decreto-ley`, `Resolución`.

Cuatro (`Decreto Foral Legislativo`, `Ley Foral`, `Orden` y `Resolución`) solo aparecen en otros resultados de esas
búsquedas, no en las diez normas del manifiesto. Entran porque la tarea y `TestEsquemaDeNormas/rangos-grabados` (T010)
exigen el conjunto exacto de las búsquedas grabadas. Las diez normas dan `Ley`, y además `Real Decreto Legislativo`
(`BOE-A-2004-4214` y `BOE-A-2015-11430`) y `Constitución` (`BOE-A-1978-31229`), todos dentro del `enum`.

## Verificación

- La prueba permanente del esquema es de T010, y esta tarea no puede añadir ningún otro fichero. El rojo → verde se
  hizo con una sonda de un solo uso en `internal/skills` que no se commitea y se borró antes de `make ci`. En rojo,
  sin el esquema, falla al abrir el fichero. En verde:
  - el esquema compila con `skills.CompilarEsquema`;
  - el `$id` es el del contrato;
  - el `enum` no repite valores y coincide con los rangos grabados, leídos desde Go;
  - un documento válido pasa `skills.ValidarDocumentoYAML`;
  - se rechazan 16 inválidos: sin `normas`, `normas` vacío, una clave de más en la raíz, `vertical`, sin título, título
    vacío, sin rango, un rango no grabado, sin materias, materias vacía, materias repetidas, una materia vacía, una
    abreviatura vacía, un identificador `BOE-B-…`, un número de diez cifras y una norma que no es un objeto.
- `jq` compara aparte, byte a byte, el `enum` con la unión de lo grabado: son iguales.
