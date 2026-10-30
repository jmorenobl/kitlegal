# Lo administrativo

El paquete de lo administrativo en `plazos-y-territorio`: el procedimiento común y el acceso a la información ante
cualquier administración. Se lee en toda pregunta de plazos o de recursos y al nombrar una norma (pasos 5 y 7 a 10
de `SKILL.md`). Las tablas las genera `make skills-sync` desde `data/`; lo demás es el detalle que solo pide esta
materia. Las líneas que da la respuesta son las de «Cómo se presenta un plazo» del cuerpo.

## Lo que la ley regula aparte

- Si el asunto es de una fila de esta tabla, la respuesta lleva la línea de su fila, entera, copiada de la última
  columna. No pidas `plazos calcular` ni des la línea del plazo ni ninguna fecha: el plazo del procedimiento común no
  es el suyo.
- Tras esa línea no expongas la norma que rige el procedimiento —sus plazos, sus recursos—: la nombra la línea.
- La línea no afirma qué dice ningún precepto, sino qué no calcula kitlegal: por eso va aunque falle una lectura. La
  lectura fallida se dice aparte, sin lo que dice el bloque. Con las cuotas de la Seguridad Social y la lectura de
  `daprimera` fallida: la línea de su fila y «no he podido leer el bloque daprimera de la LPAC: la fuente no estaba
  disponible».
- Si la conversación no dice la materia, no cae en ninguna fila. Al responder, di que la fecha no vale si el asunto es
  de los de esta tabla, nombrándolos solo con ella.

<!-- inicio de cobertura_por_materia: generada desde data/cobertura.yaml y data/normas.yaml con make skills-sync, no editar -->
| Procedimiento | Línea que da la respuesta |
|---|---|
| aplicación de los tributos en materia tributaria y aduanera y su revisión en vía administrativa | ⚠ NO CUBIERTO: aplicación de los tributos en materia tributaria y aduanera y su revisión en vía administrativa se rige por Ley 58/2003, General Tributaria (BOE-A-2003-23186), no por el procedimiento común: kitlegal no calcula sus plazos ni sus recursos (DA 1ª.2.a LPAC [BOE-A-2015-10565, bloque daprimera]). |
| gestión, inspección, liquidación, recaudación, impugnación y revisión en materia de Seguridad Social y desempleo | ⚠ NO CUBIERTO: gestión, inspección, liquidación, recaudación, impugnación y revisión en materia de Seguridad Social y desempleo lo regula aparte la ley, no el procedimiento común: kitlegal no calcula sus plazos ni sus recursos (DA 1ª.2.b LPAC [BOE-A-2015-10565, bloque daprimera]). |
| procedimientos sancionadores en materia tributaria y aduanera | ⚠ NO CUBIERTO: procedimientos sancionadores en materia tributaria y aduanera lo regula aparte la ley, no el procedimiento común: kitlegal no calcula sus plazos ni sus recursos (DA 1ª.2.c LPAC [BOE-A-2015-10565, bloque daprimera]). |
| procedimientos sancionadores en el orden social | ⚠ NO CUBIERTO: procedimientos sancionadores en el orden social lo regula aparte la ley, no el procedimiento común: kitlegal no calcula sus plazos ni sus recursos (DA 1ª.2.c LPAC [BOE-A-2015-10565, bloque daprimera]). |
| procedimientos sancionadores en materia de tráfico y seguridad vial | ⚠ NO CUBIERTO: procedimientos sancionadores en materia de tráfico y seguridad vial se rige por texto refundido de la Ley sobre Tráfico, Circulación de Vehículos a Motor y Seguridad Vial (BOE-A-2015-11722), no por el procedimiento común: kitlegal no calcula sus plazos ni sus recursos (DA 1ª.2.c LPAC [BOE-A-2015-10565, bloque daprimera]). |
| procedimientos sancionadores en materia de extranjería | ⚠ NO CUBIERTO: procedimientos sancionadores en materia de extranjería lo regula aparte la ley, no el procedimiento común: kitlegal no calcula sus plazos ni sus recursos (DA 1ª.2.c LPAC [BOE-A-2015-10565, bloque daprimera]). |
| actuaciones y procedimientos en materia de extranjería y asilo | ⚠ NO CUBIERTO: actuaciones y procedimientos en materia de extranjería y asilo lo regula aparte la ley, no el procedimiento común: kitlegal no calcula sus plazos ni sus recursos (DA 1ª.2.d LPAC [BOE-A-2015-10565, bloque daprimera]). |
| procedimientos de impugnación que una ley establece en sustitución del recurso de alzada | ⚠ NO CUBIERTO: procedimientos de impugnación que una ley establece en sustitución del recurso de alzada lo regula aparte la ley, no el procedimiento común: kitlegal no calcula sus plazos ni sus recursos (art. 112.2 LPAC [BOE-A-2015-10565, bloque a112]). |
| reclamaciones económico-administrativas | ⚠ NO CUBIERTO: reclamaciones económico-administrativas lo regula aparte la ley, no el procedimiento común: kitlegal no calcula sus plazos ni sus recursos (art. 112.4 LPAC [BOE-A-2015-10565, bloque a112]). |
<!-- fin de cobertura_por_materia -->

## Recursos y plazos

El tipo que se pasa a `plazos calcular` es el de su código, y la línea del plazo lo nombra como su primera columna.

<!-- inicio de recursos_y_plazos: generada desde data/plazos.yaml y data/normas.yaml con make skills-sync, no editar -->
| Tipo | Código | Duración | Cómputo | Precepto |
|---|---|---|---|---|
| recurso de alzada | `alzada` | 1 mes | `lpac` | art. 122.1 LPAC [BOE-A-2015-10565, bloque a122] |
| recurso potestativo de reposición | `reposicion` | 1 mes | `lpac` | art. 124.1 LPAC [BOE-A-2015-10565, bloque a124] |
| recurso contencioso-administrativo | `contencioso` | 2 meses | `contencioso` | art. 46.1 LJCA [BOE-A-1998-16718, bloque a46]; art. 128.2 LJCA [BOE-A-1998-16718, bloque a128] |
| … las demás filas de `data/plazos.yaml`, con la misma forma: las nueve de `c991808` | | | | |
<!-- fin de recursos_y_plazos -->

## El acto presunto

No hay fecha, y no se pide ningún plazo. Cita los arts. 122.1 y 124.1 LPAC y el 46.1 LJCA, leídos con
`kitlegal boe articulo`, con su cita.

## El silencio

- La fecha de inicio es la del acuerdo de iniciación o la de entrada de la solicitud en el registro.
- El plazo máximo para resolver lo fija la norma del procedimiento: léela con `kitlegal boe articulo` en la fecha de
  inicio, como cada precepto del paso 10, y pásalo a `plazos` con `--duracion` y `--unidad`. Solo si no fija ninguno,
  los tres meses del art. 21.3 LPAC, y dilo citándolo.
- `plazos` no dice si el silencio es estimatorio o desestimatorio: eso lo dice la norma, no la fecha.

## La reclamación del art. 24.2 LTAIBG

Señala como variación autonómica ante quién se reclama: el Consejo de Transparencia y Buen Gobierno o el órgano que
la comunidad haya designado (art. 24.6).

## Las leyes vertebrales

<!-- inicio de leyes_vertebrales: generada desde data/normas.yaml con make skills-sync, no editar -->
| Norma | Identificador | Rango | … las columnas y las quince filas de `leyes_vertebrales.md` de `c991808` |
|---|---|---|---|
<!-- fin de leyes_vertebrales -->
