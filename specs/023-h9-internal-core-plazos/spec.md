# Feature Specification: H9 · `internal/core/plazos` + applet `plazos` + skill `legal-core` v1

**Feature Branch**: `023-h9-internal-core-plazos`

**Created**: 2026-10-18

**Status**: Draft

**Input**: Sección «#### H9 · `internal/core/plazos` + applet `plazos` + skill `legal-core` v1» de `docs/ROADMAP.md`, en modo desatendido (ADR 0018).

## Resumen

Quien lleva un asunto, un despacho o un ciudadano, pregunta hasta qué día puede recurrir o hasta cuándo tiene la administración para resolver. Hoy `legal-core` no lo sabe responder y, si un modelo cuenta el plazo de memoria, se equivoca justo donde más cuesta: un recurso fuera de plazo no se admite. El hito hace tres cosas:

1. **Un applet nuevo, `plazos`, con un verbo, `calcular`**, y su herramienta MCP `plazos_calcular`. Calcula la fecha con el art. 30 LPAC (o con los arts. 46.1 y 128.2 LJCA y 182 a 185 LOPJ en el contencioso), con los festivos de los municipios que cuentan, y nombra el precepto que fija el plazo. No pide nada por red.
2. **Los datos de los que calcula**, en `data/`: los festivos, derivados por código de los boletines que los declaran (ADR 0039); los tipos de plazo con su precepto; y la lista de procedimientos que la ley regula aparte (DA 1ª.2 y arts. 112.2 y 112.4 LPAC; ADR 0027).
3. **`legal-core` v1**, que sitúa el asunto con `territorio`, comprueba la cobertura por materia, pide la fecha a `plazos`, lee el precepto con `boe articulo` y responde con la fecha en una línea de forma fija. Lo miden cinco evals nuevas y dos umbrales sin modelo, con 0, que deciden.

Es la pieza que H10 (el silencio, `vencimientos`) y H11 (el plazo del escrito) dan por hecha. Las fechas exigen determinismo (principio VIII): el cálculo va al binario y la skill decide qué plazo pedir.

**Cómo se citan los requisitos de otros hitos.** Los de H5 a H23.1 se citan sin guion: «H6 FR 069», «H23.1 FR 007». Los ids con guion (FR-001, SC-001…) son siempre de este spec. `<modelo>` es el id del modelo que decide (ADR 0031), hoy `claude-sonnet-5-5`; `<modo>` es `orden` u `herramienta` (H21).

**Términos de este spec.**

- **Plazo de la LPAC**: el de un tipo que no es el contencioso (FR-008). Se cuenta con el art. 30 LPAC y con dos municipios, el del interesado y el de la sede del órgano (FR-007).
- **Plazo del contencioso**: el del tipo recurso contencioso-administrativo (FR-010). Se cuenta con un municipio, el de la sede del juzgado.
- **Municipios que cuentan**: los que recibe `plazos calcular`, identificados como en `territorio resolver` (por su nombre o por su código INE).
- **Festivo cargado**: una fecha de `data/festivos/` con su ámbito (nacional, autonómico o local), su territorio y su `source`, el identificador de la disposición que la fija.
- **Tramo del plazo**: los días desde el siguiente a la fecha de inicio hasta la fecha que se da, ambos incluidos.
- **Eval de plazo**: una eval con el esperado de un plazo (FR-051). **Eval de cobertura**: una eval con el esperado de una declaración de no cubierto (FR-051).

## Clarifications

### Session 2026-10-18

- Q: ¿De qué fuente sale, sin interpretar la línea, a qué municipio pertenece una entidad menor que el BOCM da en su propia línea («Cerceda: …», «El Espartal: …»), si la línea no lo dice, `data/territorio/` no tiene entidades menores y el fichero CKAN no es fuente (ADR 0039, punto 5)? (FR-024) → A: De ninguna. La línea cuyo nombre no casa con un municipio de la Comunidad de Madrid de `data/territorio/` queda fuera de cobertura con un motivo de falta propio, sin atribuirse a ningún municipio; el aviso de entidad menor de FR-024 no se da y va a «Fuera de alcance»; el caso de la tabla de FR-082 comprueba que ese día cuenta como hábil y sin aviso; el cotejo con el CKAN (FR-084) es por código INE contra las filas del municipio (auto: conservadora; criterio d; fuente: ADR 0039 puntos 3 y 5, constitución «Gates» capa 3, «Reglas del modo desatendido», `docs/SOURCES.md`)
- Q: ¿Cómo se comprueba sin modelo que la respuesta de una eval de cobertura «da una fecha», que hace contar la sesión en el umbral de la cobertura? (FR-054) → A: Solo por la línea del plazo en su forma fija (FR-048); una fecha suelta en el texto libre no se busca, y la prohibición de pedir `plazos calcular` (FR-042) la recoge `prohibidos` en la serie (auto: criterio c; fuente: `docs/ROADMAP.md` H9, ADR 0037 en `CLAUDE.md`, `internal/evals/formato.go` y `juzgar.go` `cumpleLoEsperado`)
- Q: ¿Cómo comprueba el job que la respuesta de una eval de plazo traslada en su forma fija los avisos que da `plazos` (festivos locales que faltan en (b), art. 183 LOPJ en (e))? (FR-041, FR-053) → A: Cada eval de plazo declara los códigos de aviso esperados; un aviso ausente hace que la sesión no cumpla lo esperado y cuenta en la regla de 2 de 3 de su serie, no en el umbral de las fechas (auto: criterio a; fuente: `docs/ROADMAP.md` H9, FR-056 y FR-057, `AvisosAusentes` en `cumpleLoEsperado`)

## Criterios del hito, literales

Transcripción literal de `docs/ROADMAP.md` §4, H9. El Objetivo, la Entrega y el Alcance se reflejan en los requisitos.

- **Controles**:
  - «`make ci` en verde, con `schema-check`, el drift de `references/` y las reglas del conjunto;»
  - «property tests: sumar N días hábiles nunca acaba en un día inhábil de ninguno de los municipios; el cálculo inverso es consistente; un plazo de meses acaba en el mismo día del mes o en el último;»
  - «una tabla de casos con la fecha esperada calculada a mano y su razonamiento: los cinco de las evals, un plazo con dos municipios cuyo festivo local solo es de uno (art. 30.6), un último día en sábado, uno que cruza el cambio de año hacia 2027 en un municipio de Madrid, uno en un municipio «no comunicado» antes de su corrección y uno en una entidad menor;»
  - «la derivación de `data/festivos/` desde lo grabado, con un test que falla si `data/festivos/` no es lo que sale de las grabaciones; cada fila con su `source`; los 179 municipios de la Comunidad de Madrid con festivos locales en 2025 y 2026 o con su motivo de falta;»
  - «un test que coteja los festivos locales de 2025 con el fichero del catálogo CKAN de `evidencias/adr-0039/` y exige que las diferencias sean exactamente las de su `LEEME.md`;»
  - «e2e del applet con `--offline` y de su herramienta MCP; golden de su `--describe`;»
  - «la sección «Controles de umbral» del plan con una fila por cada umbral nuevo y las de los que ya decidían; tests de los umbrales con sesiones sintéticas;»
  - «un mutante por control: sin el art. 30.6 (un solo municipio), sin la prórroga del art. 30.5, con agosto contado en el contencioso o con una corrección del BOCM sin aplicar, algún test falla;»
  - «`CHANGELOG.md` (*Unreleased*); y el job de evals en la propuesta de cambio.»
- **Aceptación**: «en el informe del job de cierre de `legal-core`, los umbrales nuevos llevan `decide: true` y se cumplen en los dos modos, con `sin_activar` y las duraciones cumplidos, y el veredicto es aprobado; `red` está vacío. `boe-legislacion` y `jurisprudencia`, si salen en rojo, no son motivo de este hito y quedan en el informe. El informe final da los umbrales nuevos como «comprobado por su control». Después, y fuera del run porque es humano: con la release publicada, en la app de escritorio de Claude con el plugin, «me notificaron el 9 de septiembre de 2026 una resolución del Ayuntamiento de Leganés y vivo en Leganés, ¿hasta cuándo puedo presentar el recurso de reposición?» se responde con el 13 de octubre de 2026, el art. 124.1 LPAC citado y la fiesta local del 9 de octubre como razón; y «¿qué plazo tengo para recurrir una liquidación de la Agencia Tributaria?» se responde declarando la cobertura, sin dar un mes.»
- **Al cerrar**: «hay cinco applets (`boe`, `territorio`, `graph`, `cita`, `plazos`). Momento de revisar si `internal/cli` pide refactor, no extracción. Es una nota del cierre, no un hito.»

Trazabilidad resumida (el detalle, en cada requisito):

| Criterio del hito | Dónde se cumple |
|---|---|
| Entrega 1, `plazos` y su herramienta | FR-001 a FR-013, US1, US2 |
| Entrega 2, `data/festivos/` | FR-020 a FR-028, US3 |
| Entrega 3, tipos y cobertura por materia en `data/` | FR-030 a FR-032 |
| Entrega 4, `legal-core` v1 | FR-040 a FR-048, US4, US5 |
| Entrega 5, evals y umbrales | FR-050 a FR-058, US6 |
| Entrega 6, documentos | FR-060 a FR-063 |
| Lo que no cambia y el cierre | FR-070 a FR-073 |
| Controles | FR-080 a FR-090, SC-004 a SC-013 |
| Aceptación | SC-001 a SC-003 |
| Al cerrar | FR-091 |

## Relación con hitos anteriores

Ningún spec, plan ni informe de un hito anterior se edita: son el registro de sus runs. No hay ADR nuevo: los festivos los fija el ADR 0039; la cobertura por materia, el ADR 0027; los umbrales con cero de lo que comprueba un guion, el ADR 0038.

**Sustituye** (lo que dice el hito citado deja de valer, y vale lo de este spec):

- **`umbrales` de `legal-core` vacío** (H23.1 FR 007): pasa a llevar los de FR-056.
- **La `description` y el protocolo de `legal-core` v0** (H6): la skill pasa a responder también preguntas de plazos y recursos (FR-040 a FR-047). Lo que hacía del territorio sigue igual.
- **Las reglas del conjunto de evals de `legal-core`** (H6, «citas o territorio en toda positiva»): admiten también una eval positiva con el esperado de un plazo o de una declaración de no cubierto (FR-051).

**Se queda**, sin que este spec lo vuelva a definir:

- **`territorio resolver`** (H6): cómo identifica un municipio, sus códigos de salida (2 con candidatos o con un código mal formado, 3 si no está en la relación) y lo que da en `data`.
- **Qué sesiones se miden y qué es una sesión sin medir** (H7.3 FR 040 a FR 044; H21 FR 043 y FR 044): las del modelo que decide, por modo, en las evals que activan la skill. Una sesión sin medir da su motivo «de la ejecución, no de la skill».
- **El contrato de `umbrales`** (ADR 0029): `nombre`, `descripcion`, `medida`, `total`, `comparacion`, `umbral`, `cumple` y `decide`.
- **El modelo que decide, las repeticiones (3) y la regla de 2 de 3 por eval** (ADR 0016 y 0031), y la definición del job: su tope de 352 minutos, la concurrencia de `legal-core` a dos sesiones (#157, en `main`) y sus versiones.
- **El sobre** (`ok`, `fuente`, `url`, `fecha_consulta`, `hash`, `data`) y **el contrato de resultados** (ADR 0023).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Hasta qué día puedo recurrir (Priority: P1)

Una persona a la que el Ayuntamiento de Leganés notificó una resolución el 9 de septiembre de 2026, y que vive en Leganés, pregunta hasta cuándo puede presentar el recurso de reposición. La respuesta le da el 13 de octubre de 2026, calculado por el binario, con el art. 124.1 LPAC citado y la razón: el 9 de octubre es fiesta local en Leganés, el 10 y el 11 son sábado y domingo, y el 12 es fiesta nacional.

**Why this priority**: es el objetivo del hito. Un plazo mal contado hace perder el recurso.

**Independent Test**: la eval (a) en el job de evals y su caso en la tabla de casos (FR-082), sin ninguna otra pieza de la skill.

**Acceptance Scenarios**:

1. **Dado** el tipo recurso de reposición, la fecha de inicio 2026-09-09 y Leganés como municipio del interesado y de la sede del órgano, **Cuando** se ejecuta `kitlegal plazos calcular` con `--offline`, **Entonces** sale con 0 y `data` da como fecha el 2026-10-13, el precepto art. 124.1 LPAC (`BOE-A-2015-10565`) y, entre los festivos que la prórroga saltó, el 9 de octubre como local de Leganés con su `source` y el 12 de octubre como nacional con la suya.
2. **Dada** la pregunta de la eval (a), **Cuando** responde `legal-core` en el modo orden o en el modo herramienta, **Entonces** la respuesta lleva la línea del plazo en su forma fija con el 13 de octubre de 2026, el tipo y la cita del art. 124.1 LPAC.
3. **Dada** la misma pregunta sin el municipio del interesado, **Cuando** responde `legal-core`, **Entonces** pregunta por él antes de pedir el plazo, y no usa el de quien usa el kit.

---

### User Story 2 - La fecha avisa de lo que le falta (Priority: P1)

Una persona en Tordesillas (Castilla y León), con una resolución de su ayuntamiento notificada el 9 de septiembre de 2026, pregunta lo mismo. kitlegal no tiene los festivos locales de Castilla y León: la respuesta da la fecha con los festivos nacionales y autonómicos y avisa, en forma fija, de que le faltan los locales. Nunca los da por inexistentes.

**Why this priority**: un festivo local que falta adelanta la fecha; la persona tiene que saber que la fecha es la más temprana posible y por qué.

**Independent Test**: la eval (b) y el e2e del municipio no cubierto (FR-085).

**Acceptance Scenarios**:

1. **Dado** el tipo recurso de reposición, la fecha de inicio 2026-09-09 y Tordesillas como los dos municipios, **Cuando** se ejecuta `plazos calcular`, **Entonces** sale con 0, da una fecha y lleva en `data` el aviso de cobertura de festivos que nombra Tordesillas, los locales y el año 2026, y no dice que no tenga festivos locales.
2. **Dado** un plazo de un municipio de la Comunidad de Madrid cuyo tramo entra en 2027, **Cuando** se calcula, **Entonces** el resultado usa los festivos nacionales y autonómicos de 2027 y avisa de que le faltan los locales de 2027.
3. **Dado** un plazo cuyo tramo pasa por un año anterior a 2025, **Cuando** se calcula, **Entonces** avisa de que le faltan los festivos nacionales, autonómicos y locales de ese año.

---

### User Story 3 - Los festivos salen del boletín que los declara (Priority: P1)

Quien mantiene kitlegal tiene en `data/festivos/` los festivos nacionales y autonómicos de 2025 a 2027 y los locales de la Comunidad de Madrid de 2025 y 2026, cada uno con la disposición que lo fija, sacados por código de las grabaciones del BOE y del BOCM, con las correcciones del BOCM aplicadas en el orden en que se publicaron.

**Why this priority**: sin los festivos, el cálculo no vale. El ADR 0039 muestra que el catálogo CKAN no recoge las correcciones.

**Independent Test**: el test de derivación (FR-083) y el cotejo con el fichero CKAN (FR-084).

**Acceptance Scenarios**:

1. **Dadas** las grabaciones de las tres resoluciones del BOE y de las siete disposiciones del BOCM, **Cuando** el test de derivación saca `data/festivos/` de ellas, **Entonces** el resultado es igual, byte a byte, a `data/festivos/`, y cada fila lleva su `source`.
2. **Dado** Majadahonda en 2025, **Cuando** se lee `data/festivos/`, **Entonces** su fiesta de septiembre es el 15, con `source` `BOCM-20250203-16`, y no el 14 de la resolución principal.
3. **Dado** un municipio «no comunicado» en la resolución principal de 2026 (Brunete), **Cuando** se lee `data/festivos/`, **Entonces** sus festivos de 2026 son los de la corrección que los da, con su `source`.
4. **Dados** los 179 municipios de la Comunidad de Madrid de `data/territorio/`, **Cuando** se recorren 2025 y 2026, **Entonces** cada uno tiene sus festivos locales o un motivo de falta.

---

### User Story 4 - Lo que la ley regula aparte no recibe el plazo del común (Priority: P1)

Una persona pregunta qué plazo tiene para recurrir una liquidación de la Agencia Tributaria, o una multa de tráfico. La respuesta declara, en forma fija, que kitlegal no cubre ese procedimiento, nombra la norma que lo rige (la Ley General Tributaria; el texto refundido de la Ley sobre Tráfico) y no da ni el mes del recurso de reposición de la LPAC ni ninguna fecha.

**Why this priority**: darle a un asunto tributario el plazo del común es un error con apariencia de respuesta (ADR 0027; Definition of Done §1.13).

**Independent Test**: las evals (c) y (d).

**Acceptance Scenarios**:

1. **Dada** la pregunta de la eval (c), **Cuando** responde `legal-core`, **Entonces** la respuesta lleva la declaración de no cubierto en su forma fija con la Ley General Tributaria (`BOE-A-2003-23186`) y no lleva la línea del plazo.
2. **Dada** la pregunta de la eval (d), **Cuando** responde `legal-core`, **Entonces** la respuesta lleva la declaración de no cubierto con el texto refundido de la Ley sobre Tráfico y no lleva la línea del plazo.

---

### User Story 5 - El contencioso y agosto (Priority: P2)

Un despacho pregunta hasta cuándo puede interponer un recurso contencioso-administrativo cuyo plazo de dos meses pasa por agosto. La respuesta da la fecha con agosto sin contar (art. 128.2 LJCA), el último día mirado con los inhábiles del art. 182.1 LOPJ en el municipio de la sede del juzgado y, si el tramo pasa por los días del 24 de diciembre al 6 de enero, el aviso del art. 183 LOPJ.

**Why this priority**: es el plazo que más se pierde, pero va detrás del administrativo, que es la vía previa.

**Independent Test**: la eval (e) y su caso en la tabla de casos.

**Acceptance Scenarios**:

1. **Dado** el tipo contencioso, una fecha de inicio cuyo plazo de dos meses pasa por agosto y la sede del juzgado, **Cuando** se calcula, **Entonces** la fecha es la que fija el plan con agosto sin contar, y `data` nombra los arts. 46.1 y 128.2 LJCA.
2. **Dado** un plazo del contencioso cuyo tramo pasa por algún día del 24 de diciembre al 6 de enero, **Cuando** se calcula, **Entonces** esos días se cuentan, y `data` lleva el aviso del art. 183 LOPJ en su forma fija.

---

### User Story 6 - El job sale en rojo si la respuesta no da la fecha o da la del común (Priority: P1)

El job de evals de `legal-core` cuenta, sin modelo y por modo, las sesiones de las evals de plazo cuya respuesta no da la fecha esperada en su forma fija y las de las evals de cobertura que no declaran la cobertura en su forma fija o dan una fecha. Con una sola, el job sale en rojo.

**Why this priority**: sin un umbral que decida, la regla de 2 de 3 por eval (ADR 0016) deja pasar un fallo de cada tres.

**Independent Test**: los tests de los umbrales con sesiones sintéticas (FR-087).

**Acceptance Scenarios**:

1. **Dadas** 9 sesiones medidas de las evals de plazo en un modo, una de ellas sin la línea del plazo con la fecha esperada, **Cuando** se escribe el informe, **Entonces** el umbral de las fechas de ese modo da 1 de 9, `cumple: false` y el veredicto `fallo`, con un motivo que nombra la sesión.
2. **Dadas** 6 sesiones medidas de las evals de cobertura en un modo, todas con la declaración en su forma fija y ninguna con fecha, **Cuando** se escribe el informe, **Entonces** el umbral de la cobertura de ese modo da 0 de 6, `cumple: true` y ningún motivo por él.

---

### Edge Cases

- **Dos municipios con un festivo local que solo es de uno** (art. 30.6): el día es inhábil para el cómputo (FR-007).
- **El último día cae en sábado, domingo o festivo de cualquiera de los municipios que cuentan**: se prorroga al primer día hábil en todos (FR-006).
- **No hay en el mes de vencimiento un día equivalente**: vence el último día de ese mes, y si es inhábil, se prorroga (FR-005, FR-006).
- **Un festivo autonómico que la resolución del BOE limita a una parte de la comunidad**: no se aplica a toda; si está en el tramo, el resultado lo avisa (FR-011).
- **Un municipio con entidades menores con otros festivos** (Cerceda y Mataelpino en El Boalo; El Espartal en El Vellón): los del municipio son los de su línea; la línea de una entidad menor no casa con ningún municipio de la Comunidad de Madrid de `data/territorio/`, queda fuera de cobertura con su motivo de falta y ninguna de sus fechas entra en los festivos de un municipio, ni se avisa de ellas (FR-022, FR-024).
- **Una línea del BOCM que no se lee sin interpretarla** (Valdeolmos-Alalpardo en 2025 y 2026): esa entidad queda fuera de cobertura y el plazo que la usa lo avisa (FR-023, FR-026).
- **El acto es presunto**: la skill cita los arts. 122.1 y 124.1 LPAC y el 46.1 LJCA y no da fecha (FR-043).
- **Argumentos inválidos o un municipio que no se puede identificar**: código 2 o 3, como `territorio` (FR-012).
- **Ficheros del kit manipulados a mano**: lo cubre la regla genérica, defecto `inesperado` y código 1 (ADR 0023), sin casos.

## Requirements *(mandatory)*

### Functional Requirements

#### El cómputo: `plazos calcular`

- **FR-001**: El binario MUST tener un applet `plazos` con un solo verbo, `calcular`, y el servidor MCP MUST anunciar su herramienta, `plazos_calcular`, generada desde `--describe` como las demás. Recibe el tipo de plazo, la fecha de inicio, los municipios que cuentan y, solo en los tipos que no fijan su duración (el silencio y los genéricos, FR-008), la duración. `plazos` no deduce la fecha de inicio: la da quien llama. Con el código 0, `data` MUST llevar:
  - la fecha que se da, y el tipo pedido;
  - el precepto que fija el plazo, como norma (`BOE-A-…`) y bloque, y en el contencioso también el del art. 128.2 LJCA (FR-010);
  - los municipios que cuentan, cada uno con su código INE y su nombre, como los da `territorio resolver`;
  - los festivos que el cómputo saltó (FR-003 y FR-006), cada uno con su fecha, su ámbito, su municipio si es local y su `source`;
  - los avisos (FR-010, FR-011, FR-026), cada uno con su código y su texto de forma fija;
  - hasta qué disposición llegan los datos de cada ámbito del tramo (FR-026).

  Los nombres de los argumentos, de las claves y de los códigos de aviso los fija el plan. Comprobable: el e2e y el golden de FR-085.
- **FR-002**: El cómputo MUST empezar el día siguiente a la fecha de inicio, sea la de la notificación o publicación o la del silencio (arts. 30.3 y 30.4 LPAC). Comprobable: la tabla de casos (FR-082).
- **FR-003**: En días hábiles, MUST contar solo los días que no son sábado, domingo ni festivo nacional, autonómico o local de alguno de los municipios que cuentan (arts. 30.2 y 30.6 LPAC). La fecha que se da es la del enésimo día hábil. Comprobable: los property tests (FR-081) y la tabla de casos.
- **FR-004**: En días naturales, que solo se cuentan cuando quien llama lo dice (art. 30.2, párrafo segundo): con el tipo genérico de días naturales o con el silencio en días naturales (FR-009), MUST contar todos los días, y el último inhábil se prorroga como en FR-006. Comprobable: la tabla de casos.
- **FR-005**: En meses, MUST contar de fecha a fecha: el plazo vence el día del mes de vencimiento que corresponde a la fecha de inicio, y el último día de ese mes si no lo tiene (art. 30.4 LPAC). Con la fecha de inicio 2026-09-09 y un mes, el día es el 2026-10-09. Comprobable: los property tests y la tabla de casos.
- **FR-006**: Si el último día es inhábil en alguno de los municipios que cuentan, MUST prorrogarse al primer día hábil siguiente en todos (art. 30.5 LPAC; en el contencioso, el art. 185.2 LOPJ con los inhábiles de FR-010). Comprobable: la tabla de casos y el mutante de FR-089.
- **FR-007**: Un plazo de la LPAC MUST recibir exactamente dos municipios, el del interesado y el de la sede del órgano, que pueden ser el mismo. Un día inhábil en cualquiera de los dos es inhábil para el cómputo (art. 30.6 LPAC). Con uno solo, o con más de dos, sale con 2 (FR-012). Comprobable: la tabla de casos y el mutante de FR-089.
- **FR-008**: Los tipos, cada uno con su precepto en `data/` (FR-030), que el resultado nombra (FR-001), MUST ser estos y ninguno más:
  - recurso de alzada, un mes (art. 122.1 LPAC);
  - recurso potestativo de reposición, un mes (art. 124.1 LPAC);
  - recurso contencioso-administrativo, dos meses (art. 46.1 LJCA), con FR-010;
  - plazo para resolver una solicitud de acceso a la información, un mes (art. 20.1 LTAIBG);
  - plazo para reclamar contra su resolución, un mes (art. 24.2 LTAIBG);
  - el día en que se produce el silencio (FR-009);
  - los genéricos en días hábiles, en días naturales y en meses, con la duración que da quien llama y el apartado del art. 30 LPAC que los cuenta.

  `plazos` no dice si el silencio es estimatorio o desestimatorio, ni clasifica materias: calcula el tipo que se le pide. Comprobable: la tabla de casos y el golden de `--describe`.
- **FR-009**: El tipo del silencio MUST contar el plazo máximo para resolver que da quien llama, en meses, en días hábiles o en días naturales cuando quien llama lo dice, como el resto del cómputo (FR-003 a FR-005), desde la fecha del acuerdo de iniciación o de la entrada de la solicitud en el registro del órgano competente (art. 21.2 y 21.3 LPAC), y dar el día en que se produce el silencio. El plan fija, citándolo, desde qué día se cuenta y cómo se llega del vencimiento del plazo máximo al día del silencio. Comprobable: un caso de la tabla de casos con el razonamiento del plan.
- **FR-010**: El plazo del contencioso MUST:
  - recibir exactamente un municipio, el de la sede del juzgado; con otro número, sale con 2;
  - no contar agosto (art. 128.2 LJCA), con el método de descuento en un plazo de meses que fija el plan, citándolo;
  - mirar el último día con los inhábiles del art. 182.1 LOPJ, tal como los enumera su texto vigente, que el plan lee de lo grabado y cita, entre ellos los festivos a efectos laborales de la comunidad y de la localidad de la sede del juzgado, y prorrogarlo al primer día hábil siguiente (art. 185.2 LOPJ);
  - contar los días del 24 de diciembre al 6 de enero, que el art. 183 LOPJ declara inhábiles para las actuaciones judiciales y que el art. 128.2 LJCA no menciona; si el tramo pasa por alguno, `data` lleva el aviso del art. 183 LOPJ en su forma fija (FR-048).

  La fecha que se da es la más temprana de las posibles, la que no hace perder un recurso. Comprobable: la tabla de casos y el mutante de FR-089.
- **FR-011**: Un festivo autonómico que la resolución del BOE limita a una parte de la comunidad (una isla, una comarca) MUST NOT aplicarse a toda la comunidad: el día cuenta como hábil. Si está en el tramo de un plazo de un municipio de esa comunidad, `data` lleva un aviso que nombra el día y la parte. Comprobable: un test con un festivo así de lo grabado.
- **FR-012**: Los códigos de salida MUST ser los de la tabla del ADR 0023:
  - 0 con la fecha, lleve o no avisos;
  - 2 (`argumentos`) con un tipo que no es de FR-008, una fecha de inicio que no es una fecha, una duración que falta, no es un entero positivo o se da a un tipo que la fija, una duración del silencio sin su unidad (FR-009), un número de municipios distinto del de FR-007 o FR-010, o un municipio que `territorio resolver` daría con el código 2: con candidatos, `data` los da como los da `territorio`, junto al municipio pedido al que pertenecen;
  - 3 (`no-encontrado`) con un municipio que no está en la relación.

  Con varios defectos en una llamada, MUST decidir este orden: primero el tipo, la fecha, la duración y el número de municipios, con 2; luego los municipios, con 2 si alguno daría el 2 en `territorio`, con los candidatos de cada uno; y 3 solo si todo lo demás es válido y algún municipio no está en la relación. `plazos` nunca sale con 4, 5 ni 6. Comprobable: el e2e de FR-085.
- **FR-013**: `plazos` MUST NOT pedir nada por red, ni en ejecución ni con `--offline`, y MUST dar con `--offline` el mismo resultado. MUST NOT emitir operaciones de grafo: no observa identidades del mundo, y lo que nombra sale de `data/` (ADR 0014). Con la misma entrada y el mismo binario, `data` y `hash` son iguales en cada ejecución. Comprobable: el e2e de FR-085, con un test de que no hay operaciones de grafo.

#### Los festivos: `data/festivos/` (ADR 0039)

- **FR-020**: `data/festivos/` MUST salir por código de lo que graba `grabar_datos` en una tarea `[datos]` en `evidencias/h9/`:
  - los festivos nacionales y autonómicos de las resoluciones de la Dirección General de Trabajo en el BOE de 2025 (`BOE-A-2024-21316`), 2026 (`BOE-A-2025-21667`) y 2027 (`BOE-A-2026-21254`), pedidas a `/diario_boe/txt.php?id=…`, de todas las comunidades y ciudades autónomas que trae cada resolución;
  - los locales de la Comunidad de Madrid de las disposiciones del BOCM de 2025 (`BOCM-20241213-23`, `BOCM-20241227-32`, `BOCM-20250203-16` y `BOCM-20250217-24`) y de 2026 (`BOCM-20251212-34`, `BOCM-20251229-19` y `BOCM-20260108-18`), pedidas en XML.

  Cada fila MUST llevar su `source`, el identificador de la disposición que fija esa fecha. Ningún fichero de `data/festivos/` se escribe a mano. **Control**: el test de derivación de FR-083 falla en `make ci`.
- **FR-021**: Las disposiciones del BOCM de un año MUST aplicarse en el orden en que se publicaron: la que modifica gana a lo que modifica, y la fila lleva el `source` de la última que fija la fecha. Un municipio «no comunicado» no tiene festivos locales de ese año hasta que una corrección los da. Comprobable: FR-083 y el mutante de FR-089.
- **FR-022**: Los municipios MUST cruzarse con `data/territorio/` por su código INE, no por el nombre, y cada uno de los 179 municipios de la Comunidad de Madrid MUST tener en 2025 y en 2026 sus festivos locales o un motivo de falta, de un vocabulario cerrado que fija el plan. Una línea del BOCM cuyo nombre no casa con ninguno de los municipios de la Comunidad de Madrid de `data/territorio/` (hoy, las entidades menores Cerceda, Mataelpino y El Espartal) MUST quedar fuera de cobertura con un motivo de falta propio de ese vocabulario, sin atribuirse a ningún municipio: el cruce se limita a la comunidad del boletín, porque `data/territorio/municipios.yaml` tiene otra Cerceda en A Coruña (`15024`). Ninguna línea de lo grabado se pierde sin rastro. **Control**: un test de `make ci` falla si alguno no tiene ni lo uno ni lo otro, o si una línea de lo grabado no da ni festivos de un municipio ni un motivo de falta (FR-083).
- **FR-023**: Una línea que no se lee sin interpretarla MUST dejar esa entidad fuera de cobertura, con su motivo de falta: hoy, «Valdeolmos-Alalpardo: 15 de mayo y 11 de agosto (Valdeolmos) y 25 de agosto (Alalpardo)» en 2025, y la de 2026 con la misma forma. Ninguna fecha suya entra en `data/festivos/`. Comprobable: FR-083.
- **FR-024**: Los festivos locales de un municipio MUST ser los de su línea. Los de una entidad menor que el BOCM da en su propia línea («Cerceda: …», «Mataelpino: …», «El Espartal: …») MUST NOT entrar en los de ningún municipio, y el resultado MUST NOT llevar ningún aviso por ellos: ninguna fuente que se pueda grabar o leer en este hito dice, sin interpretar, a qué municipio pertenece la entidad (FR-022). Comprobable: el caso de la entidad menor de la tabla de casos (FR-082).
- **FR-025**: Un festivo autonómico limitado a una parte de la comunidad MUST quedar en `data/festivos/` como tal, con la parte que nombra la resolución, para que FR-011 lo distinga. Comprobable: FR-083.
- **FR-026**: Un plazo cuyo tramo pasa por un año sin festivos nacionales y autonómicos cargados (antes de 2025 o después de 2027), por un año sin festivos locales cargados del municipio (2027 en la Comunidad de Madrid), por un municipio sin festivos locales (fuera de la Comunidad de Madrid) o por uno con un motivo de falta (FR-022, FR-023) MUST llevar en `data` un aviso de cobertura de festivos en su forma fija (FR-048), que nombra el ámbito, el municipio si es local y los años que faltan. El cómputo sigue con lo que hay y nunca da esos festivos por inexistentes. El aviso solo se da cuando el tramo pasa por lo que falta. El resultado MUST decir hasta qué disposición llegan los datos de cada ámbito del tramo (FR-001). Comprobable: la tabla de casos y el e2e del municipio no cubierto (FR-085).
- **FR-027**: Lo que se graba MUST ser solo de fuentes con fila revisada en `docs/SOURCES.md` de `main`, al ritmo de su fila: para `data/festivos/`, las dos de festivos (`boe.fiestas-laborales` y `bocm.fiestas-locales`, 10 s en el BOCM), con su material en `evidencias/h9/` y su huella; y los preceptos que necesitan FR-010, FR-030, FR-031 y FR-052, de `boe.legislacion-consolidada`. Comprobable: el paso `grabar_datos` del workflow.
- **FR-028**: Ningún applet MUST pedir los festivos en ejecución (ADR 0017, punto 1; ADR 0039, punto 2), y `scripts/verify-sources.sh` sigue sin vigilarlos. Comprobable: FR-013.

#### Los tipos y la cobertura por materia en `data/`

- **FR-030**: `data/` MUST llevar los tipos de FR-008, cada uno con su duración cuando la fija, su cómputo (el de la LPAC o el del contencioso) y el precepto que la fija, como norma y bloque. Cada norma y cada bloque que nombra MUST estar en lo grabado del BOE, y la duración que le atribuye, en el texto grabado de ese bloque. **Control**: un test de `make ci` falla si un bloque no está grabado o si su texto no da esa duración. El plan fija cómo.
- **FR-031**: `data/` MUST llevar la lista de los procedimientos que la ley regula aparte: cada actuación o procedimiento que enumeran la DA 1ª.2 LPAC y sus arts. 112.2 y 112.4, con el precepto que lo funda y, en las entradas en que el plan la fija citando de lo grabado el precepto que la hace aplicable, la norma que lo rige por su identificador `BOE-A-…`. Lleva al menos la aplicación de los tributos y su revisión (DA 1ª.2.a), con la Ley General Tributaria (`BOE-A-2003-23186`); los procedimientos sancionadores en materia de tráfico y seguridad vial (DA 1ª.2.c), con el texto refundido de la Ley sobre Tráfico; y las reclamaciones económico-administrativas (art. 112.4). Una entrada sin norma no la toma de ningún otro sitio. Cada identificador MUST comprobarse contra lo grabado de `boe.legislacion-consolidada`, como los de `data/normas.yaml`. **Control**: un test de `make ci` falla si un identificador o un bloque no están en lo grabado.
- **FR-032**: `legal-core` MUST recibir, generadas desde `data/` con su cabecera de «no editar», `references/recursos_y_plazos.md`, con los tipos de FR-030, y la lista de cobertura por materia de FR-031, con el nombre que fija el plan. **Control**: `make skills-check`, en `make ci`, falla si difieren de lo generado.

#### La skill: `legal-core` v1

- **FR-040**: Ante una pregunta de plazos o de recursos, `skills/legal-core/SKILL.md` MUST seguir este orden:
  1. comprueba la cobertura por materia con la lista de FR-032, antes de preguntar los municipios o la fecha y antes de pedir ningún plazo; si el asunto está en ella, lo declara (FR-042) sin preguntarlos;
  2. sitúa el asunto con `territorio`. Pide los municipios que cuentan (en un plazo de la LPAC, el del interesado y el de la sede del órgano; en el contencioso, el de la sede del juzgado) y, si la persona no los ha dado, los pregunta. Nunca supone el de quien usa el kit (ADR 0027);
  3. pide el plazo a `plazos`, con la fecha de inicio que da la persona;
  4. lee con `boe articulo` el precepto que lo fija antes de citarlo.

  Comprobable: lo leen los jueces de la revisión final; su efecto lo miden las evals (FR-050).
- **FR-041**: La respuesta MUST dar la fecha que devuelve `plazos`, y ninguna otra, en la línea del plazo en su forma fija (FR-048), con su tipo y su cita. MUST trasladar cada aviso del resultado en su forma fija y decir qué festivos movieron la fecha, con lo que da `data`. Hasta qué disposición llegan los datos (FR-001) lo dice solo junto a un aviso de cobertura de festivos, del ámbito de ese aviso; sin él, no lo copia. El modelo no calcula ni corrige ninguna fecha. Comprobable: el umbral de FR-053 para la fecha, la serie de cada eval de plazo para los avisos (FR-051, FR-057) y, para la disposición, los jueces de la revisión final.
- **FR-042**: Si el asunto cae en un procedimiento de la lista de FR-031, la respuesta MUST declararlo no cubierto en la forma fija de FR-048, nombrar la norma que lo rige si la entrada la trae y, si no, el precepto de la LPAC que lo excluye, sin nombrar ninguna norma que no venga de la lista, y no dar el plazo del común: ni la línea del plazo, ni ninguna fecha. MUST NOT pedir a `plazos` el plazo del común. Comprobable: el umbral de FR-054 y la serie de cada eval de cobertura.
- **FR-043**: Ante un acto presunto, la skill MUST citar los arts. 122.1 y 124.1 LPAC y el 46.1 LJCA y MUST NOT dar fecha. Comprobable: lo leen los jueces de la revisión final.
- **FR-044**: En la reclamación del art. 24.2 LTAIBG, la skill MUST señalar como variación autonómica el órgano ante el que se reclama: el Consejo de Transparencia y Buen Gobierno, o el que la comunidad haya atribuido (art. 24.6). En el silencio, MUST decidir el plazo máximo que pide, y solo pide los tres meses del art. 21.3 LPAC cuando la norma del procedimiento no fija otro, citándolo. Comprobable: lo leen los jueces de la revisión final.
- **FR-045**: La `description` de `legal-core` MUST ganar las preguntas de plazos y de recursos de un asunto (hasta cuándo se puede recurrir, cuándo vence un plazo, cuándo se produce el silencio), para que la skill se active con ellas, y MUST seguir ganando las del territorio. **Control**: `sin_activar:<modelo>:<modo>` de `legal-core` (FR-056), con 0, que decide.
- **FR-046**: `SKILL.md` MUST seguir por debajo de 300 líneas, con la tabla de comandos de `territorio` y de `plazos` generada desde `--describe` y `kitlegal-applets` con los dos. **Control**: `make skills-check`, en `make ci`, falla con 300 líneas o más, o con una tabla distinta de la regenerada.
- **FR-047**: Las reglas de `legal-core` v0 MUST seguir valiendo para los plazos. Sin herramienta y sin binario, la respuesta no da ninguna fecha ni ningún plazo y lleva la línea `⚠ SIN CONSULTA AL BOE:` de la regla 7. Comprobable: la eval 04, que no cambia (FR-050).
- **FR-048**: Las tres formas fijas MUST tener el texto que fija el plan, de modo que un guion las compruebe sin el juicio de ningún modelo, como los avisos de vigencia de H5.1:
  - la línea del plazo, con su fecha, su tipo y su cita;
  - la declaración de que el asunto no está cubierto, con la norma que lo rige o, si la entrada no la trae, el precepto de la LPAC que lo excluye (FR-042);
  - los avisos de cobertura de los festivos y del art. 183 LOPJ.

  `SKILL.md` las lleva tal cual en bloques `text`. Comprobable: los umbrales de FR-053 y FR-054.

#### Las evals y los umbrales

- **FR-050**: Las cuatro evals de `legal-core` MUST seguir como están, y MUST entrar cinco que llaman a `plazos`, y ninguna más:
  - (a) un recurso de reposición contra una resolución del Ayuntamiento de Leganés notificada el 9 de septiembre de 2026, con el interesado en Leganés, que vence el 13 de octubre de 2026: el 9 de octubre es fiesta local en Leganés, el 10 y el 11 son sábado y domingo, y el 12, fiesta nacional;
  - (b) el mismo plazo ante el Ayuntamiento de Tordesillas, con el interesado en Tordesillas, que da su fecha con el aviso de que faltan los festivos locales;
  - (c) un recurso contra una liquidación de la Agencia Tributaria, que declara la cobertura y nombra la Ley General Tributaria;
  - (d) un recurso contra una multa de tráfico, que la declara y nombra el texto refundido de la Ley sobre Tráfico;
  - (e) un recurso contencioso-administrativo cuyo plazo pasa por agosto.

  (a), (b) y (e) son evals de plazo; (c) y (d), de cobertura. La pregunta de (e), su fecha de inicio, su sede del juzgado y su fecha esperada los fija el plan. Con nueve evals, el peor caso del trabajo de `legal-core` a dos sesiones es de 14 629 s. **Control**: `TestDefinicionDelJob`, en `make ci`, recalcula el peor caso con las evals del repositorio y falla si el tope de 352 minutos no lo cubre.
- **FR-051**: El formato de las evals MUST ganar dos esperados, que se comprueban sin modelo:
  - el de un plazo: la fecha, calculada a mano e independiente del código, con su razonamiento en el propio caso. La respuesta la tiene que llevar en la línea del plazo en su forma fija. Cada eval de plazo declara además los códigos de aviso que `plazos` da con esa entrada, también calculados a mano, y la respuesta tiene que trasladar cada uno en su forma fija: el de cobertura de los festivos locales de Tordesillas de 2026 en (b) y el del art. 183 LOPJ en (e) solo si el tramo que fija el plan pasa por algún día del 24 de diciembre al 6 de enero; en (a), ninguno. Si esos códigos van dentro del esperado del plazo o en la clave `avisos`, ampliada con los códigos de `plazos`, lo decide el plan;
  - el de una declaración de no cubierto: la norma que lo rige, por su identificador. La respuesta tiene que llevar la declaración en su forma fija con esa norma.

  Sus claves las fija el plan. `schemas/eval.yaml.json` y las reglas del conjunto de `legal-core` las admiten en una eval positiva, como hoy `citas` o `territorio`. **Control**: la validación de las evals contra el esquema, en `make ci`.
- **FR-052**: Lo que las sesiones de las cinco evals leen del BOE MUST estar grabado para el job, como lo que leen las de hoy, y ninguna sesión llega a la red. **Control**: el job de evals sale en rojo si `red` no está vacío.
- **FR-053**: Por cada modo del plan, `umbrales` de `legal-core` MUST llevar un umbral de las fechas, con el contrato del ADR 0029:
  - `total`: las sesiones de ese modo que cuentan hoy en el `total` de `sin_activar:<modelo>:<modo>` y son de una eval de plazo (9 por modo con 3 repeticiones);
  - `medida`: las de esas sesiones cuya respuesta no lleva la línea del plazo en su forma fija con la fecha esperada, cada sesión una vez como mucho;
  - `comparacion` `"<="`, `umbral` 0 y `decide: true`.

  Su nombre lo fija el plan. **Control**: con una sesión que cuenta, el veredicto es `fallo` y el job de evals de `legal-core` sale en rojo.
- **FR-054**: Por cada modo del plan, `umbrales` de `legal-core` MUST llevar un umbral de la cobertura:
  - `total`: las sesiones de ese modo que cuentan hoy en el `total` de `sin_activar:<modelo>:<modo>` y son de una eval de cobertura (6 por modo);
  - `medida`: las de esas sesiones cuya respuesta no lleva la declaración de no cubierto en su forma fija con la norma esperada, o lleva la línea del plazo en su forma fija, con la fecha que sea, cada sesión una vez como mucho. «Da una fecha» se mide solo por esa línea: una fecha suelta en el texto libre no se busca, ni cuenta una ejecución de `plazos calcular` (FR-042 la prohíbe y lo recoge la serie, FR-057);
  - `comparacion` `"<="`, `umbral` 0 y `decide: true`.

  Su nombre lo fija el plan. **Control**: con una sesión que cuenta, el veredicto es `fallo` y el job de evals de `legal-core` sale en rojo.
- **FR-055**: Los umbrales de FR-053 y FR-054 MUST:
  - poner el veredicto en `fallo` cuando no se cumplen, con un motivo que nombra el umbral con su medida y su condición, como el de todo umbral que decide, y detrás cada sesión que cuenta, por su nombre y en orden de sesión. El texto lo fija el plan. Uno que se cumple no da motivo;
  - dejar fuera de su `total` y de su `medida` las sesiones sin medir, que siguen dando su motivo «de la ejecución, no de la skill», las de un modelo informativo (hoy `claude-haiku-4-5-20251001`) y las de la eval sin binario ni servidor;
  - salir de las claves de cada eval, no de una lista en el código, y sin ningún caso por skill;
  - medir el umbral de las fechas solo por la línea del plazo con la fecha esperada: los avisos no entran en él, sino en la serie (FR-057).

  Comprobable: FR-087.
- **FR-056**: Con el plan del job, `umbrales` de `legal-core` MUST llevar exactamente `sin_activar:<modelo>:<modo>`, el umbral de FR-053 y el de FR-054 de cada modo, todos con `decide: true`, y ninguno más: `legal-core` no tiene juez ni objetivo de duración. Su orden lo fija el plan. `umbrales` de `boe-legislacion` MUST seguir con sus trece elementos, nueve que deciden, y el de `jurisprudencia`, con sus quince, once que deciden. **Control**: un test de `make ci` falla con otros elementos, otro orden o otro `decide` en alguno (FR-087).
- **FR-057**: Las evals de plazo y de cobertura MUST seguir además la regla de 2 de 3 de su serie, con sus comandos esperados y sus citas, como las de hoy. Los comandos y citas de cada una los fija el plan. En las evals de plazo, un aviso esperado que la respuesta no traslada en su forma fija, y en las de cobertura, la ejecución de `plazos calcular` o de su herramienta `plazos_calcular`, que el plan pone en `prohibidos`, hacen que la sesión no cumpla lo esperado y cuenten en su serie; ni uno ni otro entran en ningún umbral. Comprobable: el job de evals.
- **FR-058**: El hito MUST NOT cambiar `.github/workflows/evals.yml` ni los tests que leen su concurrencia: la de `legal-core` a dos sesiones está en `main` (#157). Comprobable: el diff del hito.

#### Documentos

- **FR-060**: `CHANGELOG.md` (*Unreleased*) MUST registrar el applet `plazos` y su herramienta, `data/festivos/` con sus fuentes y su cobertura, los tipos y la lista de cobertura por materia, `legal-core` v1, las cinco evals y los dos umbrales nuevos (qué cuentan, sobre qué sesiones, con 0 y que deciden) y que `umbrales` de `legal-core` deja de estar vacío. Las secciones de las versiones publicadas MUST NOT cambiar. Comprobable: SC-013.
- **FR-061**: `README.md`, `CONTRIBUTING.md` y `docs/WORKFLOW.md` MUST corregirse donde lo que dicen de `legal-core`, de sus evals, de su concurrencia o de sus umbrales quede falso o incompleto con el hito, y solo ahí. Hoy, entre otros, `CONTRIBUTING.md` da a `legal-core` una concurrencia de `1`, cuatro evals y 42 sesiones, `umbrales` vacío y unas reglas de su conjunto sin evals de plazo ni de cobertura; y su tabla del formato de eval no tiene los esperados de FR-051. Comprobable: SC-013.
- **FR-062**: Esa corrección MUST NOT reescribir lo que sigue siendo cierto, ni prometer ninguna release, ni añadir un test que ate ese texto al código. Cada cifra sale del árbol y no de memoria. Comprobable: SC-013.
- **FR-063**: MUST NOT tocarse `docs/ROADMAP.md`, ningún ADR, `docs/USO.md`, `docs/HISTORIAL.md`, la web, ni el spec, el plan o los informes de un hito anterior. Comprobable: el diff del hito.

#### Lo que no cambia y el cierre

- **FR-070**: `boe-legislacion` y `jurisprudencia` MUST NOT tocarse: ni sus `SKILL.md`, ni sus `references/`, ni sus evals, ni sus umbrales. Comprobable: el diff del hito no toca `skills/boe-legislacion/`, `skills/jurisprudencia/` ni sus evals.
- **FR-071**: Ninguna corrección ni reparación del run MUST cumplir un umbral cambiando su valor, dejándolo en `decide: false`, sacando sesiones de su `total` o cambiando las evals (ADR 0029). Si el cierre sale en rojo por un umbral de `boe-legislacion` o de `jurisprudencia`, ninguna reparación las toca ni toca sus umbrales: queda en el informe final para la persona (ADR 0038, punto 8; enmienda del 2026-10-15). Un rojo por un defecto de lo que el hito entrega sí se repara. Comprobable: lo leen los jueces de la revisión final contra el diff de cada reparación.
- **FR-072**: Dentro del run, ninguna sesión MUST abrir una sesión de evals ni un sondeo (ADR 0032): los tests usan sesiones sintéticas. Comprobable: lo leen los jueces de la revisión final contra los tests de FR-087.
- **FR-073**: Ningún municipio MUST tener un caso especial en código, `data/` ni la skill: Leganés, Tordesillas y los demás solo aparecen en fixtures, evals y la tabla de casos (Definition of Done §1.11; principio IX). Comprobable: la revisión final.

#### Controles y Definition of Done

- **FR-080**: `make ci` MUST quedar en verde, con `schema-check`, el drift de `references/` y las reglas del conjunto.
- **FR-081**: Property tests de `plazos` MUST comprobar, sobre fechas y municipios generados:
  - que sumar N días hábiles nunca acaba en un día inhábil de ninguno de los municipios;
  - que el cálculo inverso es consistente: entre el día siguiente a la fecha de inicio y la fecha que se da hay exactamente N días hábiles, y la fecha que se da es hábil;
  - que un plazo de meses acaba en el mismo día del mes o en el último, o en el hábil siguiente si ese es inhábil.
- **FR-082**: Una tabla de casos MUST llevar, cada uno con la fecha esperada calculada a mano y su razonamiento: los cinco de las evals; un plazo con dos municipios cuyo festivo local solo es de uno (art. 30.6); un último día en sábado; uno que cruza el cambio de año hacia 2027 en un municipio de Madrid; uno en un municipio «no comunicado» antes de su corrección, con los datos de la resolución principal sin la corrección; uno en una entidad menor (un plazo de El Boalo o de El Vellón cuyo tramo pasa por un festivo de Cerceda, Mataelpino o El Espartal, que no es del municipio: ese día cuenta como hábil y el resultado no lleva ningún aviso por él, FR-024); y uno del silencio (FR-009). Las fechas concretas, salvo la de (a), las fija el plan.
- **FR-083**: Un test MUST derivar `data/festivos/` desde lo grabado y fallar si no es igual; MUST comprobar que cada fila lleva su `source`, y que cada uno de los 179 municipios de la Comunidad de Madrid tiene festivos locales en 2025 y en 2026 o su motivo de falta, y que cada línea de lo grabado da festivos de un municipio o un motivo de falta (FR-022).
- **FR-084**: Un test MUST cotejar los festivos locales de 2025 de `data/festivos/` con el fichero del catálogo CKAN de `evidencias/adr-0039/`, por código INE, contra las filas del municipio (`entidad_codigo` 0; las filas de entidad del fichero no tienen código INE propio y el fichero no atribuye ninguna entidad, ADR 0039, punto 5), y exigir que las diferencias sean exactamente las de su `LEEME.md`: los seis municipios que faltan en el fichero (Belmonte de Tajo, Cabanillas de la Sierra, La Hiruela, Montejo de la Sierra, Patones y Talamanca de Jarama), las dos correcciones que no aplica (Daganzo de Arriba y Majadahonda), la fecha de Fuente el Saz de Jarama y la lectura de la línea de Valdeolmos-Alalpardo. Falla con cualquier otra diferencia o con una de esas que falte.
- **FR-085**: MUST haber e2e del applet con `--offline` y de su herramienta MCP, con la matriz territorial: un municipio con festivos locales (Leganés) y uno sin ellos (Tordesillas), con la cobertura declarada en `data` (Definition of Done §1.11); con los códigos de FR-012; y un golden de su `--describe`. Las listas literales de applets y de herramientas de los tests que ya hay se actualizan con `plazos` y `plazos_calcular`, como H23 hizo con `cita`.
- **FR-086**: La sección «Controles de umbral» del plan MUST tener una fila, en `evals:legal-core:<nombre>`, por cada umbral de FR-053 y FR-054 y por cada `sin_activar:<modelo>:<modo>` de `legal-core`, que deciden desde este hito; y las de los que ya decidían en `boe-legislacion` y `jurisprudencia`, que no cambian. MUST NOT tener fila un umbral que no decide. Comprobable: SC-002, que el informe final decide sin modelo con esas filas.
- **FR-087**: Los tests de los umbrales con sesiones sintéticas MUST cubrir: una sesión de una eval de plazo sin la línea con la fecha esperada cuenta y da `fallo`; con ella, no cuenta; una sesión de una eval de cobertura sin la declaración, o con la línea del plazo en su forma fija (FR-054), cuenta y da `fallo`; con la declaración y sin esa línea, no cuenta, aunque lleve una fecha suelta en el texto; una sesión de una eval de plazo con la línea y sin un aviso esperado no cuenta en el umbral de las fechas y no cumple lo esperado de su serie; una eval que no es de plazo ni de cobertura no entra en ningún `total`; una sesión sin medir no cuenta y da su motivo de la ejecución; una sesión de Haiku no entra. Y un test MUST comprobar los elementos de `umbrales` de las tres skills (FR-056).
- **FR-088**: El hito MUST cumplir la Definition of Done §1.13: las evals (c) y (d) son de un procedimiento que la ley regula aparte, declarado no cubierto, con la norma que lo rige y sin el plazo del común, y la (a), del procedimiento común. Comprobable: los umbrales de FR-053 y FR-054.
- **FR-089**: Un mutante por control. Con cada uno, por separado, algún test MUST fallar, y ningún mutante queda en el árbol: sin el art. 30.6 (un solo municipio); sin la prórroga del art. 30.5; con agosto contado en el contencioso; con una corrección del BOCM sin aplicar; con la comparación de FR-053 o de FR-054 en «ninguna sesión pasa» en lugar de «cada sesión»; y con el `decide` de cada uno en falso. La tarea que lo comprueba dice qué test falla con cada uno.
- **FR-090**: `CHANGELOG.md` (*Unreleased*) MUST llevar lo de FR-060, y el job de evals MUST ejecutarse en la propuesta de cambio (lo hace el workflow tras la revisión final).
- **FR-091**: El informe final MUST llevar la nota del cierre: hay cinco applets (`boe`, `territorio`, `graph`, `cita`, `plazos`), y es el momento de revisar si `internal/cli` pide refactor, no extracción. Es una nota, no trabajo del hito.

### Key Entities

- **Festivo cargado**: fecha, ámbito (nacional, autonómico, local), territorio (la comunidad, o el municipio por su código INE), la parte de la comunidad si la resolución lo limita, y `source`, la disposición que lo fija.
- **Motivo de falta**: por qué un municipio de la Comunidad de Madrid no tiene festivos locales de un año (FR-022, FR-023).
- **Tipo de plazo**: su duración, si la fija; su cómputo, el de la LPAC o el del contencioso; y su precepto, norma y bloque.
- **Procedimiento regulado aparte**: lo que enumera la DA 1ª.2 o los arts. 112.2 y 112.4 LPAC, con su precepto y, si el plan la fija (FR-031), la norma que lo rige.
- **Resultado de `plazos calcular`**: la fecha, el tipo, el precepto, los municipios, los festivos saltados, los avisos y hasta qué disposición llegan los datos (FR-001).
- **Umbral de las fechas** y **umbral de la cobertura** de `legal-core`, por modo (FR-053, FR-054).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: En el informe del job de cierre de `legal-core`, los umbrales nuevos (FR-053, FR-054) llevan `decide: true` y se cumplen en los dos modos, 0 de 9 y 0 de 6, con `sin_activar` cumplido en los dos y el trabajo terminado dentro de su tope de 352 minutos, y el veredicto es aprobado; `red` está vacío. `boe-legislacion` y `jurisprudencia`, si salen en rojo, no son motivo de este hito y quedan en el informe (FR-071). **Control**: el job de evals de `legal-core` sale en rojo, con el veredicto `fallo`, si una sesión cuenta en un umbral nuevo, si `sin_activar` no se cumple, si una sesión queda sin medir, si una serie que decide no pasa o si llega una petición a la red; el cierre del workflow cuenta cada rojo.
- **SC-002**: El informe final da los umbrales nuevos como «comprobado por su control». **Control**: lo decide sin modelo el informe final del workflow, solo para un umbral con fila en «Controles de umbral» del plan (FR-086) que el job publica, cumple y hace fallar el job.
- **SC-003**: Después del run, y fuera de él porque es humano, con la release publicada y en la app de escritorio de Claude con el plugin: «me notificaron el 9 de septiembre de 2026 una resolución del Ayuntamiento de Leganés y vivo en Leganés, ¿hasta cuándo puedo presentar el recurso de reposición?» se responde con el 13 de octubre de 2026, el art. 124.1 LPAC citado y la fiesta local del 9 de octubre como razón; y «¿qué plazo tengo para recurrir una liquidación de la Agencia Tributaria?» se responde declarando la cobertura, sin dar un mes. Lo mide una persona: no tiene control en el run.
- **SC-004**: Los tres property tests de FR-081 pasan con todas sus entradas generadas, 0 contraejemplos. **Control**: fallan en `make ci`.
- **SC-005**: Los casos de la tabla de FR-082, al menos 11, dan su fecha esperada y sus avisos: 0 diferencias. **Control**: el test de la tabla falla en `make ci`.
- **SC-006**: `data/festivos/` es igual a lo que sale de las grabaciones, 0 filas sin `source`, y 179 de 179 municipios de la Comunidad de Madrid con festivos locales o motivo de falta en 2025 y en 2026. **Control**: el test de FR-083 falla en `make ci`.
- **SC-007**: El cotejo con el fichero CKAN da exactamente las diferencias de `LEEME.md`, ninguna más ni ninguna menos. **Control**: el test de FR-084 falla en `make ci`.
- **SC-008**: Con cada uno de los mutantes de FR-089 falla al menos un test: todos de todos. **Control**: los tests de FR-081 a FR-087, en `make ci`; lo comprueban la tarea de FR-089 y la revisión final.
- **SC-009**: `plazos` hace 0 peticiones a la red y 0 operaciones de grafo, y da el mismo `data` y el mismo `hash` con `--offline` que sin él. **Control**: el e2e de FR-085 falla en `make ci`.
- **SC-010**: `skills/legal-core/SKILL.md` tiene menos de 300 líneas, y su tabla de comandos y sus `references/` son las regeneradas. **Control**: `make skills-check`, en `make ci` (FR-032, FR-046).
- **SC-011**: `umbrales` lleva en `legal-core` exactamente los seis de FR-056, los seis con `decide: true`; trece con nueve que deciden en `boe-legislacion`, y quince con once en `jurisprudencia`. **Control**: el test de FR-087 falla en `make ci`.
- **SC-012**: Los tests de los umbrales con sesiones sintéticas cubren los siete casos de FR-087: 7 de 7. **Control**: fallan en `make ci`.
- **SC-013**: `README.md`, `CONTRIBUTING.md`, `docs/WORKFLOW.md` y *Unreleased* de `CHANGELOG.md` llevan 0 afirmaciones sobre `legal-core`, sus evals, su concurrencia o sus umbrales que el hito deje falsas, y cada cifra, fichero, test y umbral que nombran existe en el árbol. **Control**: la revisión final, por lectura contra el árbol; ningún test de `make ci` ata ese texto al código (FR-062).

## Uso, de fuera adentro

Cada salida del hito, desde quien la consume (criterio de uso, ADR 0028). Lo que gana la skill es dar la fecha de un plazo calculada por el binario, con su precepto y sus razones, y no dar el plazo del común a lo que la ley regula aparte. Nada de lo que entrega crece con el uso del kit (cientos de normas y miles de bloques consultados en meses): `plazos` lee solo `data/`, que va dentro del binario, sin caché ni grafo; lo acota la pregunta, no lo acumulado. Las cifras de tamaño son estimaciones: en esta sesión no hay binario con `plazos` que medir.

| Salida | Quién la pide, cuántas veces y qué hace con ella | Tamaño | Cuándo deja de darse cada señal |
|---|---|---|---|
| El resultado de `plazos calcular` (FR-001) | `legal-core`, en el paso 3 de FR-040: una vez por plazo que pide la pregunta, casi siempre uno; dos si pide, por ejemplo, la reposición y el contencioso. Da la fecha en la línea del plazo, cita el precepto tras leerlo con `boe articulo`, explica los festivos saltados y traslada cada aviso; hasta qué disposición llegan los datos lo dice solo junto a un aviso de cobertura de festivos (FR-041). | El sobre y la fecha, el tipo, el precepto y dos municipios: unos 0,8 KB. Cada festivo saltado, unos 150 bytes: en un plazo de un mes, de 0 a 3. Como mucho, los festivos cargados del tramo de los municipios que cuentan, y solo los hay en 2025 a 2027: unos 16 por año, unos 50 en total, unos 7,5 KB. Cada aviso, unos 200 bytes: como mucho uno por ámbito al que le faltan festivos (tres: el nacional y autonómico y los dos municipios), uno del art. 183 LOPJ, y uno por cada festivo de una parte de la comunidad que esté en el tramo, que solo hay en los años cargados (en una comunidad con un festivo por isla, unos 10 por año): unos 40 en el peor caso, 8 KB. En el peor caso, menos de 17 KB por llamada; en un plazo de un mes, 1 a 2 KB. | Se calcula de nuevo en cada llamada. Cada aviso solo se da si el tramo pasa por lo que lo causa (FR-010, FR-011, FR-026). El de los festivos locales de 2027 deja de darse cuando una release lleve los del BOCM de 2027 (la tarea `[datos]` repetida). Hasta qué disposición llegan los datos es igual en cada llamada hasta que se repite la tarea `[datos]`: por eso la respuesta solo lo da con un aviso de cobertura, y deja de darlo cuando el tramo ya no pasa por lo que falta. |
| `references/recursos_y_plazos.md` y la lista de cobertura (FR-032) | `legal-core`, una vez por conversación con una pregunta de plazos: la lista, en el paso 2, para decidir si el asunto está cubierto; los tipos, para elegir el que pide. | Nueve tipos y unas seis entradas, unos 2 KB cada una. Fijo: no depende del uso. | No dan señales: cambian con `data/`. |
| La línea del plazo, la declaración de no cubierto y los avisos en la respuesta (FR-048) | La persona, una vez por respuesta. El job, que las comprueba sin modelo. | Una línea cada una. | Una por respuesta; no se acumulan. |
| `data/festivos/` (FR-020) | Solo `plazos`, dentro del binario. Ninguna skill lo lee. | Unas 250 filas nacionales y autonómicas y unas 720 locales. Fijo hasta que se repita la tarea `[datos]`. | No da señales. |
| Los umbrales de FR-053 y FR-054 en `umbrales` y su motivo de `fallo` (FR-055) | El job, una vez, que decide con ellos el veredicto; el informe final del workflow, que los lee sin modelo (SC-002); la persona y la reparación del cierre. | Seis elementos en `legal-core`, de unos 300 bytes cada uno. Cada motivo, unos 90 bytes y unos 55 por sesión que nombra, acotado por las 9 o las 6 sesiones del modo: menos de 0,6 KB. | Cada job los mide de nuevo; el `cumple: false` y su motivo dejan de darse en el primer job en que ninguna sesión cuenta. |
| `SKILL.md` de `legal-core` (FR-040 a FR-048) | El agente que activa la skill, una vez por conversación. | Menos de 300 líneas (190 en `main`). | No da señales. |

## Fuera de alcance

Del hito, literal:

- «los plazos por horas (art. 30.1 LPAC), la ampliación (art. 32), la suspensión (art. 22) y la tramitación de urgencia (art. 33);»
- «el plazo para recurrir un acto presunto: la skill cita los arts. 122.1 y 124.1 LPAC y el 46.1 LJCA y no da fecha;»
- «el procedimiento para la protección de los derechos fundamentales, en el que agosto es hábil (art. 128.2 LJCA), y los plazos de otros órdenes jurisdiccionales;»
- «los días que un calendario de días inhábiles del art. 30.7 LPAC declare sin ser festivos laborales;»
- «los festivos locales fuera de la Comunidad de Madrid, que llegan con cada comunidad desde el backlog («territorios»), y los de 2027, que llegan repitiendo la tarea `[datos]` cuando el BOCM los publique;»
- «un juez con modelo para `legal-core`, que pide rúbrica, casos etiquetados y medida (ADR 0037);»
- «que la skill use la búsqueda web del agente, que pide su propia decisión;»
- «extraer `internal/cli`.»

De lo que el hito no especifica (constitución, «Criterio de decisión autónoma», punto 2):

- Un verbo de `plazos` distinto de `calcular`, como el cálculo inverso o el del silencio de H10 (`vencimientos`): el cálculo inverso es solo una propiedad que comprueban los tests (FR-081).
- El aviso de entidad menor (que un plazo avise de un festivo de Cerceda, Mataelpino o El Espartal que no es del municipio): ninguna fuente que se pueda grabar o leer en este hito dice, sin interpretar, a qué municipio pertenece la entidad. Llega cuando una persona revise en `main` una fila de `docs/SOURCES.md` para una fuente que dé esa relación (REL, Nomenclátor del INE), no antes (ADR 0018, ADR 0039 punto 5; FR-024).
- Que `plazos` diga si el silencio es estimatorio o desestimatorio, o que clasifique la materia de un asunto.
- Más tipos de plazo que los de FR-008, y más evals que las cinco de FR-050.
- Un objetivo de duración para `legal-core`, o cualquier otro cambio de `.github/workflows/evals.yml` (FR-058).
- Cambiar las `instructions` del servidor MCP.
- Festivos de otros años, otras disposiciones que las diez de FR-020, y un buscador de correcciones del BOCM: la lista la fija una persona antes del run (ADR 0039).
- Que `plazos` guarde nada en la caché o en el grafo, o que lea la configuración de un asunto (`.kitlegal/config.yaml`) para saber los municipios: los recibe de quien llama.
- La viñeta de H9 de «Estado del repositorio» de `CLAUDE.md` y su frase del siguiente hito: las escribe la persona al fusionar.
- Editar el spec, el plan o los informes de un hito anterior; un ADR nuevo; y la release, que decide una persona.

## Assumptions

- **Lo que H6, H7.3, H21, H23 y H23.1 ya definen vale sin volver a escribirse** («Relación con hitos anteriores», «Se queda»).
- **«Las duraciones cumplidos»** de la Aceptación: `legal-core` no tiene objetivo de duración (`objetivo_de_duracion: 0` en `.github/workflows/evals.yml`, que el hito no toca, FR-058), así que su informe no lleva ningún umbral de la duración. Se lee como que el trabajo de `legal-core` termina dentro de su tope de 352 minutos (SC-001), que `TestDefinicionDelJob` recalcula con las evals del repositorio (FR-050).
- **`sin_activar` en `legal-core`**: hoy su `umbrales` está vacío porque la skill no tiene juez, ni sentencias, ni cotejo, ni objetivo de duración (`internal/evals/umbrales.go`, `umbralesDelInforme`, leído en esta sesión). Con umbrales de sus respuestas, el informe da también `sin_activar:<modelo>:<modo>` de cada modo, como en las otras dos skills; la Aceptación lo nombra.
- **Las sesiones de los umbrales**: 3 evals de plazo y 2 de cobertura, con 3 repeticiones, dan 9 y 6 sesiones por modo (FR-053, FR-054). En el último cierre versionado de `legal-core` (`specs/022-h23-1-jurisprudencia-coteja-el/gates/evals/legal-core.json`, leído con `jq` en esta sesión) `repeticiones` es 3, el modelo informativo es `claude-haiku-4-5-20251001`, `umbrales` está vacío y el veredicto es aprobado.
- **Los esperados nuevos del formato de eval** (FR-051): el hito pide el esperado de un plazo; el umbral de la cobertura necesita saber qué evals son de cobertura y qué norma esperan, y ninguna clave de hoy lo dice (`internal/evals/formato.go`, leído en esta sesión). Por eso entra también el esperado de la declaración de no cubierto, que sirve a un umbral que pide el hito (ADR 0029).
- **Cómo lee `legal-core` el precepto** (FR-040, paso 4): el hito pide «lee con `boe articulo`»; la v0 delega el texto en `boe-legislacion` y no lleva sus órdenes. Si la orden la ejecuta la propia skill o la delega, y qué skills no se activan en las evals nuevas, lo decide el plan; las evals miden que la orden se ejecutó.
- **El art. 182.1 LOPJ** (FR-010): el hito resume sus inhábiles entre guiones; se leen todos los que enumera su texto vigente, que el plan transcribe de lo grabado.
- **El caso «no comunicado antes de su corrección»** (FR-082): el cálculo no depende de la fecha en que se pide, así que el caso se lee con los datos que daría la resolución principal sin la corrección; con ellos, el plazo avisa de que le faltan los festivos locales de ese municipio.
- **Los municipios se identifican como en `territorio`** (FR-012): por nombre o por código INE, con sus mismos códigos de salida y sus candidatos.
- **Las listas literales de los tests que ya hay** (FR-085): `internal/app/testdata/script/h21-mcp-herramientas.txtar`, `internal/empaquetado/ejecutar_test.go` y `snapshot_piezas_test.go` nombran hoy las herramientas una a una (buscado con `grep` en esta sesión). Con `plazos_calcular` cambia el guion, como cambió con H23; las dos de `internal/empaquetado` no, porque piden que estén, no que sean las únicas (research V11).
- **Lo comprobado en esta sesión**, sin ejecutar ningún test ni `make ci`: en `BOCM-20251212-34` (el XML de `evidencias/adr-0039/bocm/`), «Leganés: 14 de agosto y 9 de octubre.», catorce municipios «no comunicado», Cerceda, Mataelpino y El Espartal en su propia línea y la línea de Valdeolmos-Alalpardo con paréntesis; `data/territorio/municipios.yaml` tiene 179 municipios de la provincia 28, entre ellos Leganés (`28074`) y «Boalo, El» (`28023`), y Tordesillas es `47165`, de Castilla y León; el fichero CKAN da a Cerceda y Mataelpino como entidades de El Boalo y a El Espartal como de El Vellón; `data/normas.yaml` tiene la LGT (`BOE-A-2003-23186`), la LJCA, la LOPJ y la LTAIBG, y no tiene el texto refundido de la Ley sobre Tráfico. Las fechas de 2026 del calendario (el 9 de octubre, viernes; el 12, lunes) son del hito.
- **Los nombres técnicos que aparecen** (`plazos`, `calcular`, `plazos_calcular`, `data/festivos/`, `references/recursos_y_plazos.md`, `umbrales`, `sin_activar`, `TestDefinicionDelJob`, `make skills-check`, `schemas/eval.yaml.json`) existen en el repositorio o los fija el hito. Quedan para el plan los nombres de los argumentos, de las claves de `data`, de los códigos de aviso, de las claves de las evals y de los umbrales nuevos, el texto de las formas fijas, el orden de `umbrales` y las fechas de los casos de la tabla salvo la de (a).
