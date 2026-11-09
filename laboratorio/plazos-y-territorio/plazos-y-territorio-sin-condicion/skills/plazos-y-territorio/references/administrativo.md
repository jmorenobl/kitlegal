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
| plazo para resolver una solicitud de acceso a la información pública | `acceso-resolver` | 1 mes | `lpac` | art. 20.1 LTAIBG [BOE-A-2013-12887, bloque a20] |
| reclamación contra la resolución de una solicitud de acceso | `acceso-reclamar` | 1 mes | `lpac` | art. 24.2 LTAIBG [BOE-A-2013-12887, bloque a24] |
| día en que se produce el silencio administrativo | `silencio` | la que recibe, con su unidad | `lpac` | art. 21.3 LPAC [BOE-A-2015-10565, bloque a21]; art. 24.4 LPAC [BOE-A-2015-10565, bloque a24] |
| plazo genérico | `generico` | la que recibe, en `meses` | `lpac` | art. 30.4 LPAC [BOE-A-2015-10565, bloque a30] |
| plazo genérico | `generico` | la que recibe, en `dias-habiles` | `lpac` | art. 30.2 LPAC [BOE-A-2015-10565, bloque a30] |
| plazo genérico | `generico` | la que recibe, en `dias-naturales` | `lpac` | art. 30.2 LPAC [BOE-A-2015-10565, bloque a30] |
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
| Norma | Abreviatura | Identificador | Rango | Materias |
|---|---|---|---|---|
| Real Decreto de 24 de julio de 1889 por el que se publica el Código Civil. | CC | `BOE-A-1889-4763` | Real Decreto | derecho civil, fuentes del derecho |
| Constitución Española. | CE | `BOE-A-1978-31229` | Constitución | organización territorial, autonomía local |
| Ley 7/1985, de 2 de abril, Reguladora de las Bases del Régimen Local | LRBRL | `BOE-A-1985-5392` | Ley | régimen local |
| Ley Orgánica 6/1985, de 1 de julio, del Poder Judicial. | LOPJ | `BOE-A-1985-12666` | Ley Orgánica | poder judicial |
| Ley 29/1998, de 13 de julio, reguladora de la Jurisdicción Contencioso-administrativa. | LJCA | `BOE-A-1998-16718` | Ley | jurisdicción contencioso-administrativa, recurso contencioso-administrativo |
| Ley 1/2000, de 7 de enero, de Enjuiciamiento Civil. | LEC | `BOE-A-2000-323` | Ley | proceso civil |
| Ley 38/2003, de 17 de noviembre, General de Subvenciones. | LGS | `BOE-A-2003-20977` | Ley | subvenciones |
| Ley 47/2003, de 26 de noviembre, General Presupuestaria. | LGP | `BOE-A-2003-21614` | Ley | presupuestos públicos, control del gasto público |
| Ley 58/2003, de 17 de diciembre, General Tributaria. | LGT | `BOE-A-2003-23186` | Ley | tributos |
| Real Decreto Legislativo 2/2004, de 5 de marzo, por el que se aprueba el texto refundido de la Ley Reguladora de las Haciendas Locales. | TRLRHL | `BOE-A-2004-4214` | Real Decreto Legislativo | tributos, haciendas locales |
| Ley 19/2013, de 9 de diciembre, de transparencia, acceso a la información pública y buen gobierno. | LTAIBG | `BOE-A-2013-12887` | Ley | transparencia, acceso a la información pública |
| Ley 39/2015, de 1 de octubre, del Procedimiento Administrativo Común de las Administraciones Públicas. | LPAC | `BOE-A-2015-10565` | Ley | procedimiento administrativo |
| Ley 40/2015, de 1 de octubre, de Régimen Jurídico del Sector Público. | LRJSP | `BOE-A-2015-10566` | Ley | régimen jurídico del sector público, potestad sancionadora |
| Ley 9/2017, de 8 de noviembre, de Contratos del Sector Público, por la que se transponen al ordenamiento jurídico español las Directivas del Parlamento Europeo y del Consejo 2014/23/UE y 2014/24/UE, de 26 de febrero de 2014. | LCSP | `BOE-A-2017-12902` | Ley | contratación pública |
| Ley Orgánica 3/2018, de 5 de diciembre, de Protección de Datos Personales y garantía de los derechos digitales. | LOPDGDD | `BOE-A-2018-16673` | Ley Orgánica | protección de datos personales, derechos digitales |
<!-- fin de leyes_vertebrales -->
