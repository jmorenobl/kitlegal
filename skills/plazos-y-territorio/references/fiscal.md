# Lo fiscal

El paquete tributario en `plazos-y-territorio`: el recurso de reposición y la reclamación económico-administrativa
contra un acto de aplicación de los tributos del Estado, y lo tributario que la ley regula aparte. Se lee en toda
pregunta de plazos o de recursos de un tributo, una liquidación, una sanción tributaria o la Agencia Tributaria, y al
nombrar una norma del paquete (pasos 5 y 7 a 10 de `SKILL.md`). Las tablas las genera `make skills-sync` desde
`data/`; lo demás es el detalle que solo pide esta materia. Las líneas que da la respuesta son las de «Cómo se presenta
un plazo» del cuerpo.

## Lo que cubren los dos tipos

- Solo el recurso de reposición y la reclamación económico-administrativa en única o primera instancia contra un acto
  de aplicación de los tributos del Estado, lo dicte la Administración del Estado o la de una comunidad autónoma
  (art. 226.a LGT [BOE-A-2003-23186, bloque a226]). Una liquidación del IRPF o del IVA de la Agencia Tributaria es de
  estos actos.
- El inicio es el día de la notificación del acto: es la fecha que se pasa a `plazos calcular`. Si el plazo empieza
  de otro modo —una deuda de vencimiento periódico y notificación colectiva, una retención, una repercusión, una
  factura o el silencio—, no pidas ningún plazo: es de lo no declarado, con su línea, y no digas cómo se cuenta.
- Los dos municipios son el del interesado y el de la sede del órgano que dictó el acto, que es también el que recibe
  la reclamación (art. 235.3 LGT). Si la persona no dice dónde está el órgano, pregúntalo, como en el
  paso 8 del cuerpo; nunca lo supongas por el del interesado.
- Se elige una vía: la reposición es potestativa y va antes de la reclamación, y con las dos interpuestas contra el
  mismo acto se tramita la primera (art. 222.2 LGT [BOE-A-2003-23186, bloque a222]). Los dos tipos dan la misma fecha,
  y la línea del plazo nombra el de la vía que pide la persona.
- El tipo y su precepto se leen en la fecha de inicio, como cada precepto del paso 10: el 223.1 de la reposición y el
  235.1 de la reclamación.

## Lo que queda fuera

- Un tributo local —el IBI, la plusvalía, el impuesto de vehículos, una tasa municipal— no es de estos dos tipos: es de
  la fila de la aplicación de los tributos de la tabla de abajo, con su línea, aunque lo liquide un ayuntamiento de
  la conversación.
- Un tributo de una hacienda foral y un tributo propio de una comunidad autónoma son de lo no declarado: la línea de lo
  no declarado del cuerpo, sin fecha y sin nombrar la norma que los rige.
- Una sanción tributaria es de su fila de la tabla de abajo, también su recurso.

## Lo que la ley regula aparte

- Si el asunto es de una fila de esta tabla, la respuesta lleva la línea de su fila, entera, copiada de la última
  columna. No pidas `plazos calcular` ni des la línea del plazo ni ninguna fecha: el plazo de los dos tipos no es el
  suyo.
- Tras esa línea no expongas la norma que rige el procedimiento —sus plazos, sus recursos—: la nombra la línea.
- La línea no afirma qué dice ningún precepto, sino qué no calcula kitlegal: por eso va aunque falle una lectura. La
  lectura fallida se dice aparte, sin lo que dice el bloque.
- Si la conversación no dice la materia, no cae en ninguna fila.

Los motivos de la línea; el primero es el de la fila con norma:

```text
se rige por <norma> (<identificador de la norma>), no por el procedimiento común (<precepto> [<identificador>, bloque <id>])
lo regula aparte la ley, no el procedimiento común (<precepto> [<identificador>, bloque <id>])
```

<!-- inicio de cobertura_por_materia: generada desde data/cobertura.yaml y data/normas.yaml con make skills-sync, no editar -->
| Procedimiento | Línea que da la respuesta |
|---|---|
| aplicación de los tributos en materia tributaria y aduanera y su revisión en vía administrativa, salvo el recurso de reposición y la reclamación económico-administrativa en única o primera instancia contra un acto de aplicación de los tributos del Estado | ⚠ NO CUBIERTO: aplicación de los tributos en materia tributaria y aduanera y su revisión en vía administrativa, salvo el recurso de reposición y la reclamación económico-administrativa en única o primera instancia contra un acto de aplicación de los tributos del Estado: se rige por Ley 58/2003, General Tributaria (BOE-A-2003-23186), no por el procedimiento común (DA 1ª.2.a LPAC [BOE-A-2015-10565, bloque daprimera]). kitlegal no da sus plazos, órganos, vías ni escritos. |
| procedimientos sancionadores en materia tributaria | ⚠ NO CUBIERTO: procedimientos sancionadores en materia tributaria: se rige por Ley 58/2003, General Tributaria (BOE-A-2003-23186), no por el procedimiento común (DA 1ª.2.c LPAC [BOE-A-2015-10565, bloque daprimera]). kitlegal no da sus plazos, órganos, vías ni escritos. |
| reclamaciones económico-administrativas, salvo la reclamación en única o primera instancia contra un acto de aplicación de los tributos del Estado | ⚠ NO CUBIERTO: reclamaciones económico-administrativas, salvo la reclamación en única o primera instancia contra un acto de aplicación de los tributos del Estado: se rige por Ley 58/2003, General Tributaria (BOE-A-2003-23186), no por el procedimiento común (art. 112.4 LPAC [BOE-A-2015-10565, bloque a112]). kitlegal no da sus plazos, órganos, vías ni escritos. |
<!-- fin de cobertura_por_materia -->

## Recursos y plazos

El tipo que se pasa a `plazos calcular` es el de su código, y la línea del plazo lo nombra como su primera columna.

<!-- inicio de recursos_y_plazos: generada desde data/plazos.yaml y data/normas.yaml con make skills-sync, no editar -->
| Tipo | Código | Duración | Cómputo | Precepto |
|---|---|---|---|---|
| recurso de reposición contra un acto de aplicación de los tributos | `reposicion-tributaria` | 1 mes | `lpac` | art. 223.1 LGT [BOE-A-2003-23186, bloque a223]; art. 7.2 LGT [BOE-A-2003-23186, bloque a7] |
| reclamación económico-administrativa en única o primera instancia | `reclamacion-economico-administrativa` | 1 mes | `lpac` | art. 235.1 LGT [BOE-A-2003-23186, bloque a235]; art. 7.2 LGT [BOE-A-2003-23186, bloque a7] |
<!-- fin de recursos_y_plazos -->
