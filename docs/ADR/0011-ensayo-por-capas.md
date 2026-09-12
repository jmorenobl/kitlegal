# 0011 · El ensayo lo describe cada capa con efectos y lo presenta el kernel

- **Estado**: aceptada
- **Fecha**: 2026-09-12
- **Hito**: H2

## Contexto y problema

H1 fijó qué hace `--dry-run`: describir en lugar de ejecutar, terminar con código `0` y dejar la
descripción en la **salida de error, siempre visible**, sin depender del nivel del registro de eventos
(`specs/002-h1-kernel-cli-multicall/contracts/banderas-y-exit-codes.md` §3 y §6). En H1 eso era barato:
ninguna capa tenía efectos, así que la única descripción posible —qué applet, qué verbo, con qué
argumentos— la conocía el propio kernel y la escribía él.

H2 trae la primera capa con efectos, `internal/httpx`. Bajo ensayo, `Pedir` no abre conexión y devuelve
una respuesta que declara que no emitió nada (FR-050, FR-052, FR-065), pero **qué** se habría pedido
—método y dirección— solo lo sabe ahí abajo, y FR-051 exige que llegue a la salida de error igual de
visible que la línea de H1. El problema es que ninguna capa por debajo del kernel tiene forma de escribir:
R5 de `contracts/reglas-de-arquitectura.md` reserva la salida a `internal/render`, y el presentador lo
construye e inyecta la raíz de composición. ADR 0005, además, cerró el contrato del applet en
`Resultado{Procedencia, Datos}`: hoy no hay por dónde subir una descripción que no es contenido citable.

## Opciones consideradas

1. **Emitirla por `slog`**, que ya llega a todas las capas. Rechazada: el registro de eventos es filtrable
   por nivel, y un requisito de visibilidad incondicional no se implementa sobre un canal que se filtra.
   Es la misma razón por la que H1 sacó del registro la línea de `--dry-run` (H1 research.md D10).
2. **Dar al cliente un `io.Writer` al que escribir su descripción.** Rechazada por dos motivos: viola R5
   —la escritura dejaría de estar en `internal/render`— y no admite valor por omisión seguro, porque el
   cero de un `io.Writer` es `nil`; quien construyera el cliente sin pasarlo perdería la descripción en
   silencio, que es justo lo que FR-051 prohíbe.
3. **Cambiar la firma de `Ejecutar`** para que el applet reciba por dónde describir. Rechazada: obliga a
   tocar el contrato de los más de veinte applets del roadmap —y cada implementación— para transportar
   una línea de texto, cuando el valor de retorno ya viaja del applet al kernel en cada invocación.
4. **Que la descripción viaje en el valor de retorno**, que es el único camino que ya existe del applet al
   kernel y del kernel al presentador.

## Decisión

Se adopta la **opción 4**, ampliando ADR 0005 sin sustituirlo.

- `schema.Resultado` gana un tercer campo, `Ensayo []string`: una línea por operación, lo que cada capa
  con efectos habría hecho en lugar de hacerlo. Es dominio puro —un `[]string`, ninguna importación
  nueva— y solo se rellena bajo `--dry-run`.
- **Quien tiene el efecto describe**: el adaptador copia ahí `respuesta.Descripcion()` de cada petición
  que `internal/httpx` no emitió. El dominio no sabe presentarla y el cliente no sabe escribirla.
- **Quien tiene el presentador presenta**: `internal/app` escribe, detrás de la línea de H1 y por el
  presentador —nunca por `slog`—, una línea `--dry-run: se habría pedido <método> <dirección>` por
  elemento de `Ensayo`. Las dos van en un **único** aviso, de modo que ninguna escritura pueda fallar a
  medias y dejar una descripción incompleta sin que nadie lo note.
- **El código de salida no cambia**: estar en ensayo no es un fallo, así que sigue siendo `0` salvo que la
  invocación sea inválida por argumentos o configuración, que se comprueban igualmente porque no
  necesitan red (FR-065).
- `Ensayo` **no entra en el sobre ni en la huella**: no es contenido citable. El sobre conserva sus seis
  claves exactas, y bajo `--dry-run` no se emite ninguno.

## Consecuencias

- El contrato del applet de ADR 0005 pasa a ser `Resultado{Procedencia, Datos, Ensayo}`. Es
  retrocompatible: un applet que no rellene `Ensayo` se comporta exactamente como en H1, y los guiones de
  e2e de H1 siguen valiendo sin tocarlos.
- Todo applet que use una capa con efectos queda obligado a propagar la descripción hacia arriba. Si no lo
  hace, su `--dry-run` no miente —no se ejecuta nada— pero calla, y FR-051 no se cumple para esa fuente.
  El control es de cada applet: `TestDryRunPresentaElEnsayo` fija el comportamiento del kernel, y cada
  adaptador prueba el suyo.
- La regla R1 sigue intacta: `internal/core/schema` no gana ninguna importación, y la dependencia va
  `httpx → schema`, nunca al revés (FR-055).
- El mecanismo no es de HTTP: cualquier capa con efectos futura —la caché, el almacén, el grafo, la
  generación de documentos— describe por el mismo canal, sin volver a tocar el contrato.
