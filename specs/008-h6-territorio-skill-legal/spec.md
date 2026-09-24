# Feature Specification: H6 · `territorio` + skill `legal-core` v0

**Feature Branch**: `008-h6-territorio-skill-legal`

**Created**: 2026-09-20

**Status**: Draft

**Input**: Sección «#### H6 · `territorio` + skill `legal-core` v0» de `docs/ROADMAP.md`, con la enmienda del 2026-09-18 (ADR 0017), en modo desatendido.

## Resumen

H5 dejó `boe-legislacion`: una skill que sabe leer y citar cualquier norma consolidada del BOE. Lo que no sabe ninguna skill todavía es **dónde** está la pregunta. Sin territorio no hay normativa autonómica, ni boletín donde mirar, ni festivos, ni órgano competente: toda skill que quiera ser genérica para cualquier municipio de España necesita esa pieza antes de razonar (constitución, principio IX; `refs/mapa-sistema-legal-skills.md` §3, capa 0).

H6 es el primer hito de la fase 1 y entrega dos cosas:

1. **El applet `territorio`**, con el verbo `resolver`: dado un nombre de municipio o su código INE, devuelve municipio, código INE, provincia, comunidad autónoma, DIR3 del ayuntamiento, régimen (común o foral), boletines aplicables y `cobertura` —qué del territorio está configurado y qué no—. No pide nada por red: su dato vive congelado en `data/territorio/`, generado fuera de la ejecución a partir de la relación oficial del INE y del Registro de Entidades Locales (ADR 0017). Un nombre que comparten varios municipios devuelve la lista de candidatos y exit 2. Con él entran los dos primeros identificadores del proyecto, código INE y DIR3, con su formato y su dígito de control.
2. **La skill `legal-core` v0**, la skill madre: su protocolo **empieza por identificar el territorio** de la pregunta y solo después razona, con `references/leyes_vertebrales.md` y `references/jerarquia_normativa.md` generadas desde `data/`, y evals que exigen precisamente eso.

La regla que gobierna todo el hito es la de la genericidad territorial: **ningún caso especial para un municipio** en código, en `data/` ni en la skill; fuera del territorio configurado —solo la Comunidad de Madrid lo está— la salida **declara su cobertura** en lugar de fallar, de vaciarse o, peor, de inventar un boletín o un código DIR3 que nadie ha verificado.

## Clarifications

### Session 2026-09-20

- Q: ¿Qué municipios entran en el fichero de correspondencia INE→DIR3: solo los de la muestra comprobada uno a uno, o todos aquellos a los que la regla comprobada en la muestra se aplica sin discrepancia? → A: B. Entra todo municipio con fila en el REL cuyo número de inscripción es coherente con su código INE y su dígito de control; la muestra verifica la regla, no las filas, y su evidencia va al registro de verificación (FR-047). Sin fila o con número incoherente, queda fuera y se declara en `cobertura`. Si la muestra revela discrepancias más allá de casos aislados, la derivación pasa a ser tabla y se decide en la pausa humana de la tarea `[datos]`. (auto: criterio a; fuente: docs/ROADMAP.md §4 H6 Controles; ADR 0017, decisión 3 y consecuencias; spec FR-023, FR-042, FR-046 a FR-048)
- Q: ¿De qué parte de `data/` salen las dos referencias generadas de `legal-core`: cómo se distinguen las normas «vertebrales» y qué fichero sostiene `jerarquia_normativa.md`? → A: B. Campo booleano opcional `vertebral: true` en `data/normas.yaml` (con su ampliación en `schemas/normas.yaml.json`) para `leyes_vertebrales.md`; y un fichero de datos nuevo (p. ej. `data/jerarquia.yaml`) con esquema propio en `schemas/` para `jerarquia_normativa.md`. (auto: criterio c; fuente: refs/kitlegal-estructura-y-ecosistema.md §2; docs/ROADMAP.md §4 H6 Alcance; CLAUDE.md «Skills sin código»; spec FR-065 a FR-067, FR-070)
- Q: ¿Cómo se expresan y se juzgan mecánicamente las evals de `legal-core`, si el formato común solo conoce comandos y citas de norma y bloque del BOE? → A: A. El formato común se extiende en H6, compatible hacia atrás: `comandos` gana la variante `{applet: territorio, verbo: resolver, municipio: …}`; aparece un esperado propio de territorio (comunidad, provincia, boletines y aspectos de `cobertura`); y «toda eval activa exige `citas`» pasa a «exige al menos un esperado verificable», con `citas` obligatorias cuando la eval afirme contenido de norma. Cambia `schemas/eval.yaml.json` —fichero existente, bajo el régimen de FR-085— y con él el juicio mecánico que lo aplica (detalle para `plan`, no requisito: hoy vive en `internal/evals`); las evals de `boe-legislacion` no cambian. (auto: criterio c; fuente: spec FR-080 a FR-084; schemas/eval.yaml.json; docs/ROADMAP.md §4 H5; constitución, criterio 1)
- Q: ¿Cómo obtiene el binario los ficheros de `data/territorio/` en ejecución: empaquetados, leídos del sistema de ficheros o mezcla? → A: A. Empaquetados en el binario al compilar (`//go:embed`) y validados contra su esquema en test; un fichero ausente es un fallo de compilación, no hay modo de fallo en ejecución y FR-016 se mantiene. (Detalle para `plan`, no requisito: `go:embed` no sube de directorio, así que la directiva vive en un paquete que contenga la ruta `data/` —raíz del módulo—, no en `internal/core/territorio`. FR-056 solo fija el comportamiento observable.) (auto: criterio c; fuente: spec FR-009, FR-016, FR-053, SC-001; ADR 0017, consecuencias; CLAUDE.md «Multicall»)
- Q: Cuando una pregunta de `legal-core` exige el texto de una norma, ¿qué hace la skill y qué applets lleva su tabla de comandos? → A: A. `legal-core` v0 resuelve territorio y razona con sus referencias (identificadores y jerarquía, nunca texto de norma); para el texto de un artículo remite a `boe-legislacion` (delegación en un solo sentido); su tabla de comandos generada desde `--describe` lleva solo los verbos de `territorio`, y sus evals miden territorio, cobertura y no activación. (auto: criterio a; fuente: docs/ROADMAP.md §4 H6 Entrega; refs/mapa-sistema-legal-skills.md §3 capa 0; spec FR-060 a FR-068)

## Criterios del hito, literales

Transcripción literal de `docs/ROADMAP.md` §4, H6, para que la trazabilidad del spec se pueda comprobar sin volver al roadmap. El tachado y el corchete son del propio roadmap (enmienda del 2026-09-18, ADR 0017).

- **Objetivo**: «que toda skill sepa, antes de razonar, en qué territorio está la pregunta y qué normas vertebrales aplican. Es la pieza que hace genéricas a las demás.»
- **Entrega**: «`kitlegal territorio resolver <nombre|código INE>` devuelve municipio, código INE, provincia, comunidad autónoma, DIR3 del ayuntamiento, régimen (común o foral), boletines aplicables y `cobertura` (qué del territorio está configurado y qué no). Skill `legal-core` v0: protocolo que empieza por identificar el territorio, `references/leyes_vertebrales.md` y `jerarquia_normativa.md`.»
- **Alcance**: «registro de municipios desde la relación oficial del INE y ~~DIR3 desde el inventario público (formato, licencia y forma de actualización a verificar en tarea `[datos]`, con fila en `docs/SOURCES.md`)~~ DIR3 desde un fichero versionado en `data/territorio/`, generado por tarea `[datos]` a partir del INE (dígito de control) y del Registro de Entidades Locales (número de inscripción), nunca pedido en red [enmienda 2026-09-18, ADR 0017: el inventario DIR3 no es descargable de forma automatizada —el conjunto de `datos.gob.es` está despublicado (404), el fichero del CTT está tras un WAF que devuelve 0 bytes a un cliente no interactivo y los servicios web exigen Red SARA—; la única vía pública es el REL, y el código se deriva de su número de inscripción, derivación que hay que verificar antes de fijarla]; `data/territorio/` con configuración por comunidad, solo la Comunidad de Madrid rellena (su boletín, el BOCM, hace de provincial por ser uniprovincial); el resto de comunidades responde con los datos nacionales y cobertura parcial. Nombre ambiguo: lista de candidatos y exit 2. Ids que entran aquí: código INE y DIR3 (`internal/core/ids`, formato y dígito de control), que en H7 son los ids naturales de `Municipio` y `Organo`. Verificar con `boe buscar` los ids de la tabla de leyes vertebrales y fijarlos en `data/normas.yaml`.»
- **Controles**: «matriz territorial en e2e: Leganés (cubierto), Tordesillas (datos nacionales completos, boletines autonómico y provincial declarados no cubiertos, nada inventado), un municipio de Navarra o del País Vasco (régimen foral marcado), un nombre ambiguo; fuzz del parser de INE y DIR3; la tarea `[datos]` que genera la correspondencia INE→DIR3 verifica la derivación contra DIR3 real en una muestra que incluya los casos donde puede romperse (un municipio fusionado o renombrado, uno de régimen foral, uno con entidades locales menores) y deja fuera, declarándolo en `cobertura`, todo municipio cuyo código no se verifique; evals de `legal-core` que exigen identificar el territorio.»
- **Aceptación**: «en Claude Code, «¿qué comunidad, provincia y boletines corresponden a mi ayuntamiento?» se responde para cualquier municipio, con la cobertura explícita cuando no es de la Comunidad de Madrid.»

Trazabilidad resumida (el detalle, en cada requisito y criterio):

| Criterio literal | Dónde se cumple |
|---|---|
| `territorio resolver <nombre\|código INE>` devuelve municipio, INE, provincia, comunidad, DIR3, régimen, boletines y `cobertura` | FR-001 a FR-008, US1, SC-001 |
| Nombre ambiguo: lista de candidatos y exit 2 | FR-013 a FR-015, US4, SC-004 |
| registro de municipios desde la relación oficial del INE | FR-040, FR-041, US7, SC-007 |
| DIR3 desde un fichero versionado en `data/territorio/`, generado por tarea `[datos]` a partir del INE y del REL, **nunca pedido en red** | FR-042 a FR-045, FR-049, US7, SC-007, SC-012 |
| **Control**: la tarea `[datos]` verifica la derivación contra DIR3 real en una muestra con municipio fusionado o renombrado, foral y con entidades locales menores | FR-046, FR-047, US7, SC-008 |
| **Control**: deja fuera, declarándolo en `cobertura`, todo municipio cuyo código no se verifique | FR-023, FR-048, US7, SC-008 |
| `data/territorio/` con configuración por comunidad, solo la Comunidad de Madrid rellena; el BOCM hace de provincial por ser uniprovincial | FR-050 a FR-053, US1, SC-005 |
| el resto de comunidades responde con los datos nacionales y cobertura parcial | FR-020 a FR-022, FR-054, US2, SC-002 |
| régimen común o foral | FR-006, FR-055, US3, SC-003 |
| Ids que entran aquí: código INE y DIR3 (`internal/core/ids`, formato y dígito de control) | FR-030 a FR-034, US4, SC-006 |
| **Control**: fuzz del parser de INE y DIR3 | FR-035, SC-006 |
| Skill `legal-core` v0: protocolo que empieza por identificar el territorio | FR-060 a FR-064, US5, SC-009 |
| `references/leyes_vertebrales.md` y `jerarquia_normativa.md` | FR-065 a FR-067, US5, SC-009 |
| Verificar con `boe buscar` los ids de la tabla de leyes vertebrales y fijarlos en `data/normas.yaml` | FR-070 a FR-074, US6, SC-010 |
| **Control**: evals de `legal-core` que exigen identificar el territorio | FR-080 a FR-084, US5, SC-011, SC-015 |
| **Control**: matriz territorial en e2e (Leganés cubierto; Tordesillas con boletines declarados no cubiertos y nada inventado; foral; nombre ambiguo) | FR-090, FR-091, US1 a US4, SC-001 a SC-004 |
| **Aceptación**: «¿qué comunidad, provincia y boletines corresponden a mi ayuntamiento?» se responde para cualquier municipio, con la cobertura explícita fuera de la Comunidad de Madrid | FR-100 a FR-102, US5, SC-013 |
| Definition of Done §1.10 (evals escritas antes que el código **y en verde**) | FR-083, FR-084, SC-011, SC-015 |
| Definition of Done (`make ci`, esquemas, exit codes, e2e, cobertura, `CHANGELOG.md`, skills sin drift) | FR-092 a FR-099, SC-012, SC-014 |
| Constitución, capa 3 (`schemas/` y `testdata/`: material existente y material nuevo, solo en tareas `[datos]`) | FR-085, FR-086, y FR-045, FR-049 y FR-073 en su caso |

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Quien pregunta por un municipio del territorio configurado obtiene su terreno completo (Priority: P1)

Una persona en Claude Code pregunta qué comunidad, provincia y boletines corresponden a su ayuntamiento, y nombra Leganés. La skill resuelve el territorio con el binario antes de decir nada y responde con municipio, código INE, provincia, comunidad autónoma, DIR3 del ayuntamiento, régimen y los boletines que le aplican, diciendo que el territorio está configurado.

**Why this priority**: es la entrega del hito y la aceptación literal. Sin esto no hay pieza que haga genéricas a las demás skills.

**Independent Test**: se prueba sola ejecutando `territorio resolver` sobre el municipio de referencia y comprobando el sobre y el código de salida, sin ninguna otra parte del hito.

**Acceptance Scenarios**:

1. **Given** el repositorio con los datos de `data/territorio/` generados, **When** se ejecuta `territorio resolver Leganés --json`, **Then** el sobre sale con `ok` verdadero, exit 0 y `data` con nombre del municipio, código INE, provincia, comunidad autónoma, DIR3 del ayuntamiento, régimen y boletines aplicables.
2. **Given** la misma consulta, **When** se lee `cobertura` en `data`, **Then** declara configurados el boletín autonómico y el provincial —los dos, el BOCM, por ser la Comunidad de Madrid uniprovincial— y configurado el DIR3 del ayuntamiento.
3. **Given** la misma consulta hecha por código INE en lugar de por nombre, **When** se comparan las dos salidas, **Then** `data` es idéntico en las dos y su huella también.
4. **Given** la bandera `--offline`, **When** se repite la consulta, **Then** la salida es byte a byte la misma y el proceso no abre ninguna conexión.

---

### User Story 2 - El mismo comando responde fuera del territorio configurado sin inventar nada (Priority: P1)

La misma pregunta, con un municipio de una comunidad que nadie ha configurado todavía (Tordesillas, Valladolid). La respuesta trae los datos nacionales completos —municipio, INE, provincia, comunidad, régimen y, si está verificado, DIR3— y declara que el boletín autonómico y el provincial no están configurados, sin nombrar ninguno y sin dar a entender que no existan.

**Why this priority**: es la mitad del principio IX y lo que separa «genérico» de «solo Madrid». Un applet que se vaciara o fallara fuera del territorio cubierto haría inútil la skill para cualquier otro municipio.

**Independent Test**: se prueba sola con un municipio de una comunidad no configurada, comprobando que los datos nacionales están completos, que `cobertura` los declara no configurados y que en la salida no aparece el nombre de ningún boletín no configurado.

**Acceptance Scenarios**:

1. **Given** una comunidad sin configuración en `data/territorio/`, **When** se resuelve un municipio suyo, **Then** exit es 0 y `data` trae municipio, código INE, provincia, comunidad autónoma y régimen.
2. **Given** esa misma salida, **When** se leen los boletines aplicables, **Then** solo aparece el boletín estatal y `cobertura` declara no configurados el autonómico y el provincial.
3. **Given** esa misma salida, **When** se busca en ella el nombre o el código de cualquier boletín autonómico o provincial, **Then** no aparece ninguno: lo no configurado se declara, no se adivina.
4. **Given** esa misma salida, **When** una skill la lee, **Then** nada en ella permite concluir que ese municipio no tenga boletín autonómico o provincial.

---

### User Story 3 - El régimen foral queda marcado allí donde lo hay (Priority: P2)

Quien pregunta por un municipio de Navarra o del País Vasco recibe la misma respuesta que cualquier otro, con el régimen marcado como foral, aunque su comunidad no esté configurada.

**Why this priority**: el régimen cambia qué normativa aplica (tributos, haciendas locales) y una skill que lo ignore razonará mal; va después de las dos primeras porque es un campo más de la misma respuesta.

**Independent Test**: se prueba sola resolviendo un municipio de Navarra o del País Vasco y comprobando el régimen y la cobertura.

**Acceptance Scenarios**:

1. **Given** un municipio de Navarra o del País Vasco, **When** se resuelve, **Then** `data` marca régimen foral y exit es 0.
2. **Given** un municipio de cualquier otra comunidad o ciudad autónoma, **When** se resuelve, **Then** `data` marca régimen común.
3. **Given** el municipio foral, **When** se lee `cobertura`, **Then** el régimen consta como dato nacional conocido, con independencia de que su boletín autonómico no esté configurado.

---

### User Story 4 - Un nombre ambiguo devuelve los candidatos, y uno que no existe se distingue de uno mal escrito (Priority: P2)

Quien escribe un nombre que comparten varios municipios recibe la lista de candidatos con su código INE y su provincia, y un código de salida que dice que el argumento no basta. Quien escribe un municipio que no existe recibe otro código distinto.

**Why this priority**: sin distinguir «hay varios» de «no hay ninguno», la skill concluiría «no existe» a partir de una ambigüedad, que es justo lo que las reglas invariantes prohíben.

**Independent Test**: se prueba sola con un nombre compartido por varios municipios de la relación del INE, un nombre inexistente y un código mal formado.

**Acceptance Scenarios**:

1. **Given** un nombre que la relación del INE da a más de un municipio, **When** se resuelve, **Then** exit es 2 y el mensaje del sobre de fallo lista todos los candidatos con su código INE y su provincia.
2. **Given** un nombre que no corresponde a ningún municipio, **When** se resuelve, **Then** exit es 3.
3. **Given** un código INE con dígito de control incorrecto o con un formato imposible, **When** se resuelve, **Then** exit es 2 y el mensaje dice qué tiene de malo.
4. **Given** un código INE bien formado —cinco cifras con provincia entre `01` y `52` y municipio distinto de `000`— que no está en la relación, **When** se resuelve, **Then** exit es 3. (Un código con provincia fuera de ese rango no está bien formado: es entrada mal formada, exit 2.)

---

### User Story 5 - La skill `legal-core` empieza por el territorio y lo demuestra en sus evals (Priority: P1)

Una sesión de Claude Code con `legal-core` instalada responde la pregunta de la aceptación para cualquier municipio: primero resuelve el territorio con el binario, después razona con las leyes vertebrales y la jerarquía normativa, y dice explícitamente qué parte del territorio no está configurada cuando el municipio no es de la Comunidad de Madrid.

**Why this priority**: el producto son las skills (constitución, principio VIII); el applet existe para esta skill.

**Independent Test**: se prueba sola con las evals de `legal-core`, que exigen que la respuesta pase por el applet y traiga el territorio identificado.

**Acceptance Scenarios**:

1. **Given** la skill instalada, **When** se pregunta por el territorio de un municipio cubierto, **Then** la respuesta identifica municipio, provincia, comunidad y boletines, y cada dato sale de la salida del applet en esa misma conversación.
2. **Given** un municipio no cubierto, **When** se hace la misma pregunta, **Then** la respuesta dice qué está configurado y qué no, y no nombra ningún boletín que el applet no haya devuelto.
3. **Given** una pregunta que no es de territorio ni de derecho, **When** se evalúa la activación, **Then** la skill no se activa.
4. **Given** `SKILL.md`, **When** se comprueba mecánicamente, **Then** tiene frontmatter válido, menos de 300 líneas y sus `references/` coinciden con lo generado desde `data/`.

---

### User Story 6 - Los identificadores de las leyes vertebrales quedan verificados, no recordados (Priority: P3)

Quien mantiene `data/normas.yaml` cierra la tabla de leyes vertebrales: cada identificador `BOE-A-…` de la tabla queda comprobado contra una búsqueda real del BOE antes de fijarse, y las referencias generadas lo reflejan.

**Why this priority**: es un cierre que el hito arrastra desde H5 («la tabla completa se cierra en H6») y que protege a todas las skills de citar un identificador escrito de memoria; va la última porque no bloquea al applet.

**Independent Test**: se prueba sola comprobando cada identificador de la tabla contra la búsqueda grabada correspondiente, sin red.

**Acceptance Scenarios**:

1. **Given** la tabla de leyes vertebrales de `refs/mapa-sistema-legal-skills.md`, **When** se compara con `data/normas.yaml`, **Then** toda ley de la tabla está en el fichero con su identificador.
2. **Given** cada identificador nuevo, **When** corre la comprobación sin red del repositorio, **Then** coincide con el que devuelve la búsqueda grabada del BOE para el título de esa norma.
3. **Given** el fichero actualizado, **When** se regeneran las referencias, **Then** `references/` de las dos skills queda sin drift.

---

### User Story 7 - La correspondencia INE→DIR3 se genera, se verifica y se congela fuera de la ejecución (Priority: P2)

Quien prepara los datos ejecuta una tarea `[datos]`, fuera del bucle de implementación y con pausa: descarga la relación del INE y el volcado del REL, deriva el DIR3 de cada ayuntamiento, **verifica la derivación contra DIR3 real** en una muestra que incluye los casos donde puede romperse, y deja versionado en `data/territorio/` solo lo que queda verificado, con su procedencia.

**Why this priority**: es la condición para que el DIR3 del applet sea un dato y no una invención (ADR 0017, decisión 3); va después de la entrega visible porque el applet responde sin DIR3 declarándolo en `cobertura`.

**Independent Test**: se prueba sola revisando el fichero generado, su procedencia y el registro de la muestra verificada.

**Acceptance Scenarios**:

1. **Given** la tarea `[datos]`, **When** se ejecuta, **Then** produce los ficheros de `data/territorio/` a partir de las descargas públicas y ninguna petición sale del binario.
2. **Given** la muestra de verificación, **When** se revisa, **Then** incluye al menos un municipio fusionado o renombrado, uno de régimen foral y uno con entidades locales menores, y registra para cada uno el DIR3 real contra el que se comparó.
3. **Given** un municipio cuyo código no se verifica, **When** se resuelve con el applet, **Then** la salida no trae DIR3 y `cobertura` lo declara no verificado.
4. **Given** el repositorio ya generado, **When** corre `make ci`, **Then** todo pasa sin red y sin ninguna descarga.

---

### Edge Cases

- **Nombre como lo escribe el INE**: municipios con artículo pospuesto (`Coruña, A`) o con nombre bilingüe con barra (`Donostia/San Sebastián`); ¿se reconoce el nombre en la forma oficial, en la directa y en cada una de las lenguas?
- **Mayúsculas y diacríticos**: `leganes` y `LEGANÉS` deben llevar al mismo municipio; una coincidencia insensible no puede convertir dos municipios distintos en el mismo.
- **Entrada que parece código y nombre a la vez**: qué se hace con una entrada de solo cifras que no es un código válido.
- **Código con y sin dígito de control**: cinco cifras y seis cifras; con seis, el dígito se comprueba.
- **Ciudades autónomas**: Ceuta y Melilla no son comunidad ni provincia en el sentido habitual; su régimen es común y sus boletines no están configurados.
- **Comunidad uniprovincial distinta de Madrid**: no se infiere que su boletín autonómico haga de provincial; se declara no configurado.
- **Municipio sin DIR3 verificado**: nunca se devuelve un código calculado como si fuera registral (ADR 0017, decisión 3).
- **Municipio desaparecido por fusión**: el nombre antiguo no está en la relación vigente; la salida no lo inventa y el código de salida distingue «no existe» de «ambiguo».
- **Boletín que el aplicativo conoce pero cuya configuración está incompleta**: la cobertura enumera cada aspecto, nunca omite uno por no saber qué decir de él.
- **Régimen foral en un municipio de Álava, Gipuzkoa o Bizkaia frente a uno de Navarra**: los dos se marcan foral.

## Requirements *(mandatory)*

### Functional Requirements

#### El applet `territorio` y su verbo `resolver`

- **FR-001**: El binario DEBE registrar un applet `territorio` con el verbo `resolver`, invocable como `kitlegal territorio resolver <nombre|código INE>` y, por el multicall, como `territorio resolver <nombre|código INE>`.
- **FR-002**: `resolver` DEBE aceptar como único argumento posicional obligatorio el nombre del municipio o su código INE, y DEBE distinguirlos por la forma de la entrada.
- **FR-003**: La salida DEBE ser el sobre obligatorio `{ok, fuente, url, fecha_consulta, hash, data}`, con las seis claves y ninguna más (ADR 0006).
- **FR-004**: Como `territorio` no consulta ninguna fuente en ejecución, su procedencia DEBE ser la del espacio de nombres reservado (`fuente` `kitlegal.territorio`, `url` `kitlegal:applet/territorio`, ADR 0006), y NO DEBE usar el identificador de una fuente de `docs/SOURCES.md` como `fuente` del sobre.
- **FR-005**: Cada dato de `data` cuya procedencia sea un fichero congelado DEBE llevar su `source` con el identificador de la fila de `docs/SOURCES.md` que lo origina (`ine.municipios`, `ine.codigos-territoriales`, `mpt.rel`) o el de la configuración de `data/territorio/` que lo fija (ADR 0017, decisión 3).
- **FR-006**: `data` DEBE contener, para un municipio resuelto: nombre del municipio, código INE, provincia, comunidad autónoma, DIR3 del ayuntamiento, régimen (común o foral), boletines aplicables y `cobertura`.
- **FR-007**: La clave de cada campo y su forma DEBEN describirse con `--describe`, y `schemas/` DEBE contener el esquema de entrada y salida del applet, sin drift respecto de lo que `--describe` emite (régimen de FR-086).
- **FR-008**: Los boletines aplicables DEBEN darse por nivel (estatal, autonómico, provincial), incluyendo solo los que el territorio tenga configurados y siempre el estatal.
- **FR-009**: `--offline` y `--dry-run` DEBEN comportarse sin cambiar la respuesta: el applet no abre conexiones en ningún caso, de modo que `--offline` devuelve exactamente lo mismo.

#### Entrada, ambigüedad y códigos de salida

- **FR-010**: Una consulta resuelta DEBE terminar con exit 0.
- **FR-011**: Un municipio que no está en la relación DEBE terminar con exit 3 («no encontrado»). Por código, esto solo alcanza a los **bien formados** según FR-012: cinco cifras (o seis con su dígito) con provincia entre `01` y `52` y municipio distinto de `000`.
- **FR-012**: Una entrada mal formada —código con dígito de control incorrecto, con un número de cifras imposible, con la provincia fuera del rango `01`-`52`, con el municipio `000` o vacía— DEBE terminar con exit 2, con un mensaje que diga qué tiene de malo. Una entrada de solo cifras que no cumple la gramática **no llega a ser un código**, y por tanto nunca termina en 3.
- **FR-013**: Un nombre que corresponde a más de un municipio DEBE terminar con exit 2 y NO DEBE elegir uno por su cuenta.
- **FR-014**: En ese caso, el mensaje del sobre de fallo DEBE listar todos los candidatos con su código INE y su provincia, en orden de código INE. (El sobre de fallo conserva las seis claves de FR-003; lo que es exactamente `{clase, mensaje}` por ADR 0006 es su `data`, así que la lista viaja en el mensaje.)
- **FR-015**: La coincidencia por nombre DEBE ser insensible a mayúsculas y a los signos diacríticos, y DEBE reconocer el nombre tal como lo escribe la relación del INE, su forma con el artículo antepuesto y, en los nombres bilingües, cada una de sus formas; dos municipios distintos nunca pueden colapsar en uno por efecto de esa normalización.
- **FR-016**: El applet NO DEBE decidir nunca exit 4, 5 ni 6: no consulta fuentes ni cruza la frontera humana, y sus datos no pueden faltar en ejecución (FR-056). La única excepción no es del applet: si se agota `--timeout`, el kernel termina con 4 para todo applet (FR-020 de H1).

#### Cobertura y no invención

- **FR-020**: `cobertura` DEBE enumerar todos los aspectos del territorio que el applet puede llenar —al menos boletín autonómico, boletín provincial y DIR3 del ayuntamiento— y decir de cada uno si está configurado o verificado, sin omitir ninguno.
- **FR-021**: Fuera del territorio configurado, `data` NO DEBE contener el nombre, el código ni la URL de ningún boletín no configurado.
- **FR-022**: La salida NO DEBE permitir concluir que un boletín no existe: `cobertura` distingue «no configurado» de «no existe», y el segundo caso no se emite nunca.
- **FR-023**: Un municipio sin DIR3 verificado DEBE resolverse igualmente, sin campo DIR3 con valor y con `cobertura` declarándolo no verificado; NUNCA DEBE devolverse un código derivado como si fuera registral (ADR 0017, decisión 3).
- **FR-024**: Ningún municipio concreto puede aparecer como caso especial en el código ni en `data/`; los municipios solo aparecen en fixtures, e2e y evals (constitución, principio IX; Definition of Done §11).

#### Identificadores: código INE y DIR3 (`internal/core/ids`)

- **FR-030**: `internal/core/ids` DEBE ofrecer el análisis y la validación del código INE de municipio: formato, provincia y dígito de control.
- **FR-031**: `internal/core/ids` DEBE ofrecer el análisis y la validación del código DIR3 de un ayuntamiento, incluida su relación con el código INE y el dígito de control.
- **FR-032**: Los dos DEBEN normalizar su entrada de forma idempotente: normalizar lo ya normalizado no lo cambia, y analizar y volver a escribir devuelve la misma cadena normalizada.
- **FR-033**: Una entrada inválida DEBE rechazarse con un error tipado que el kernel traduzca a exit 2, nunca con `panic` ni con un valor por defecto.
- **FR-034**: El paquete NO DEBE implementar otros identificadores (ELI, ECLI, CELEX, NIF): entran con sus hitos.
- **FR-035**: DEBE haber objetivos de fuzz para los analizadores de código INE y de DIR3, con corpus mínimo versionado en `testdata/fuzz/`, que comprueben que ninguna entrada provoca `panic` y que la ida y vuelta de toda entrada aceptada es estable (régimen de FR-086).

#### Datos congelados en `data/territorio/` y la tarea `[datos]`

- **FR-040**: El registro de municipios DEBE salir de la relación oficial del INE, con código INE, dígito de control, nombre, provincia y comunidad autónoma.
- **FR-041**: Ese registro DEBE vivir versionado en `data/territorio/` en un formato de texto legible y comparable en un diff, con su procedencia y la fecha del fichero de origen.
- **FR-042**: La correspondencia INE→DIR3 DEBE vivir versionada en `data/territorio/`, generada a partir del código INE (dígito de control) y del número de inscripción del REL.
- **FR-043**: Ningún applet, ni en ejecución ni en grabación, DEBE pedir por red el inventario DIR3, el REL o la relación del INE (ADR 0017, decisión 1).
- **FR-044**: Los ficheros de `data/territorio/` y el de la jerarquía normativa (FR-067) DEBEN validarse contra un esquema propio en `schemas/`, como `data/normas.yaml` desde H5; esos esquemas nuevos entran bajo el régimen de FR-086.
- **FR-045**: La generación de esos ficheros DEBE hacerse en una tarea `[datos]` con pausa humana, fuera del bucle de implementación (constitución, capa 3 y reglas del modo desatendido).
- **FR-046**: Esa tarea `[datos]` DEBE verificar la derivación del DIR3 contra DIR3 real en una muestra que incluya al menos un municipio fusionado o renombrado, uno de régimen foral y uno con entidades locales menores.
- **FR-047**: La verificación DEBE quedar registrada: para cada municipio de la muestra, el código derivado, el código real contra el que se comparó y de dónde salió ese código real.
- **FR-048**: Todo municipio cuyo código no se verifique DEBE quedar fuera del fichero y declararse en `cobertura` (FR-023). Entra en el fichero todo municipio al que la regla de derivación se aplica sin discrepancia: fila en el REL cuyo número de inscripción es coherente con su código INE y con su dígito de control oficial. La muestra de FR-046 verifica la regla, no cada fila; su evidencia va al registro de FR-047 y no a la salida del applet. Cada fila lleva su `source` (`ine.municipios` y `mpt.rel`). Si la muestra revela discrepancias más allá de casos aislados, la derivación pasa a ser tabla (ADR 0017, consecuencias), y eso se decide con la muestra delante, dentro de la pausa humana de la tarea `[datos]`.
- **FR-049**: `docs/SOURCES.md` DEBE seguir siendo cierto respecto de los ficheros generados (fecha del fichero y formato de origen); cualquier cambio en ese documento va dentro de la tarea `[datos]`, con revisión humana, y NO DEBE añadirse ningún caso a `scripts/verify-sources.sh` para estas fuentes, que no se piden en red (ADR 0017, consecuencias).

#### Configuración por comunidad

- **FR-050**: `data/territorio/` DEBE admitir una configuración por comunidad autónoma, con los boletines aplicables y lo demás que varíe por territorio.
- **FR-051**: Solo la Comunidad de Madrid DEBE venir rellena en este hito.
- **FR-052**: En la Comunidad de Madrid, el BOCM DEBE figurar a la vez como boletín autonómico y como provincial, por ser comunidad uniprovincial, y la razón DEBE constar en la propia configuración.
- **FR-053**: Añadir un territorio nuevo DEBE consistir en añadir su configuración, sin tocar código ni skills (constitución, principio IX; ROADMAP §2, «Configuración por datos»).
- **FR-054**: Una comunidad sin configuración DEBE responder con los datos nacionales completos y cobertura parcial, nunca con un error ni con una respuesta vacía.
- **FR-055**: El régimen (común o foral) DEBE conocerse para todo municipio de España, con independencia de que su comunidad tenga configuración: es dato nacional, no configuración de territorio.
- **FR-056**: Los ficheros de `data/territorio/` DEBEN viajar dentro del binario publicado: la respuesta NO DEBE depender del directorio de trabajo, de una variable de entorno ni de ninguna ruta configurable, y DEBE ser la misma se ejecute el binario desde donde se ejecute (SC-001). La ausencia, la falta o la invalidez de uno de esos ficheros frente a su esquema DEBE detectarse antes de ejecutar —al construir el binario y en `make ci`—, nunca en ejecución, de modo que no exista modo de fallo por datos ausentes (de ahí FR-016). Refrescar los datos exige repetir la tarea `[datos]` y publicar una release. Con qué mecanismo se empaquetan y en qué paquete vive ese empaquetado lo decide `plan.md`.

#### Skill `legal-core` v0

- **FR-060**: DEBE existir `skills/legal-core/` con `SKILL.md`, `references/` generadas y `scripts/` como symlinks al binario (CLAUDE.md, «Skills sin código»).
- **FR-061**: El protocolo de `SKILL.md` DEBE empezar por identificar el territorio de la pregunta, antes de razonar sobre normas.
- **FR-062**: `SKILL.md` DEBE ordenar que el territorio se resuelva con el applet y que ningún dato de territorio se dé por sabido; y DEBE trasladar a la respuesta lo que `cobertura` declare no configurado.
- **FR-063**: `SKILL.md` DEBE recoger las reglas invariantes de todas las skills: no inventar contenido legal, citar por identificador, distinguir ley de reglamento, señalar variación autonómica, no concluir «no existe» a partir de un resultado sin cobertura completa y no ejecutar ninguna acción con identidad.
- **FR-064**: `SKILL.md` DEBE tener frontmatter válido, menos de 300 líneas y la tabla de comandos generada desde `--describe`, sin drift.
- **FR-065**: DEBE existir `skills/legal-core/references/leyes_vertebrales.md`, generada desde `data/` con la cabecera «generado …, no editar», con las leyes vertebrales y su identificador `BOE-A-…`.
- **FR-066**: DEBE existir `skills/legal-core/references/jerarquia_normativa.md`, generada desde `data/` con la misma cabecera, con la jerarquía de fuentes por nivel (UE, Estado, comunidad autónoma, provincia, municipio), qué boletín publica cada nivel y las reglas de interpretación (competencia antes que jerarquía, ley posterior, ley especial, reglamento nunca contra ley).
- **FR-067**: La única fuente de verdad de las dos referencias DEBE ser `data/`, y `make skills-sync` DEBE regenerarlas; `make ci` falla si difieren de lo commiteado. `leyes_vertebrales.md` sale de las normas de `data/normas.yaml` marcadas con un campo booleano opcional `vertebral: true` (añadido a `schemas/normas.yaml.json` bajo el régimen de FR-085); `references/normas.md` de `boe-legislacion` sigue generándose del fichero entero. `jerarquia_normativa.md` sale de un fichero de datos nuevo en `data/` (p. ej. `data/jerarquia.yaml`) con esquema propio en `schemas/`, con los niveles, la clase de boletín que publica cada uno y las reglas de interpretación.
- **FR-068**: `legal-core` NO DEBE incluir en v0 ni plazos, ni recursos, ni competencia: llegan con H9 y sus hitos.
- **FR-069**: Sus referencias llevan identificadores y jerarquía, nunca el texto de una norma. Cuando la pregunta exige el texto de un artículo, `SKILL.md` DEBE ordenar no citarlo desde la referencia ni de memoria y remitir a `boe-legislacion` (delegación en un solo sentido); citar el identificador de una norma basta con la referencia generada, pero afirmar lo que dice un artículo exige la consulta de `boe-legislacion` en esa misma conversación. La tabla de comandos de `SKILL.md` lleva únicamente los verbos de `territorio`, y `scripts/` solo su symlink; las evals de `legal-core` miden territorio, cobertura y no activación, y las de norma y bloque siguen siendo las de `boe-legislacion`.

#### Leyes vertebrales en `data/normas.yaml`

- **FR-070**: Toda ley de la tabla de leyes vertebrales de `refs/mapa-sistema-legal-skills.md` §1.4 DEBE estar en `data/normas.yaml` con su identificador `BOE-A-…`, su título, su rango, su abreviatura, sus materias y la marca `vertebral: true` (FR-067). Las normas del fichero que no están en esa tabla no llevan la marca.
- **FR-071**: Cada identificador nuevo DEBE verificarse con `boe buscar` antes de fijarse; ninguno se escribe de memoria.
- **FR-072**: La comprobación que ata cada identificador a la búsqueda del BOE DEBE correr sin red en `make ci`, sobre respuestas grabadas.
- **FR-073**: Las grabaciones que esa comprobación necesita DEBEN obtenerse en una tarea `[datos]` con pausa, nunca dentro del bucle de implementación ni de un job; si alguna de ellas modifica un fichero de grabación ya versionado, rige además FR-085.
- **FR-074**: Tras el cambio, `references/normas.md` de `boe-legislacion` y las referencias de `legal-core` DEBEN quedar regeneradas y sin drift, y el conjunto de evals de `boe-legislacion` DEBE seguir cumpliendo sus reglas.

#### Evals de `legal-core`

- **FR-080**: DEBE existir `evals/legal-core/` con evals que exigen identificar el territorio antes de responder.
- **FR-081**: El conjunto DEBE incluir al menos una eval sobre un municipio del territorio configurado y otra sobre un municipio no configurado (Definition of Done §11; `refs/kitlegal-estructura-y-ecosistema.md` §3).
- **FR-082**: El conjunto DEBE incluir al menos una eval de no activación, con una pregunta que no debe disparar la skill.
- **FR-083**: Las evals DEBEN escribirse antes que el código de la skill (constitución, principio III; ROADMAP §6.1) y DEBEN acabar en verde: la Definition of Done §1.10 exige las dos cosas, y el verde se mide en la ejecución de aceptación del hito con la regla de ADR 0016 (SC-015).
- **FR-084**: Las evals de `legal-core` DEBEN poder expresar y juzgar mecánicamente que la respuesta pasó por `territorio resolver` y trae el territorio identificado. Para ello se extiende el formato común de eval, de forma compatible hacia atrás: `comandos` admite la variante `{applet: territorio, verbo: resolver, municipio: …}`; hay un esperado propio de territorio (comunidad, provincia, boletines y aspectos de `cobertura` que la respuesta debe declarar); y la regla «toda eval que activa la skill exige `citas`» pasa a «exige al menos un esperado verificable», con `citas` obligatorias siempre que la eval afirme contenido de norma. El esquema del formato común (`schemas/eval.yaml.json`, bajo el régimen de FR-085) y el juicio mecánico que lo aplica DEBEN admitir esa variante sin dejar de aceptar lo ya escrito: las evals de `boe-legislacion` siguen siendo válidas sin modificarse y sus reglas del conjunto siguen en verde (SC-015). Qué paquete implementa el juicio lo decide `plan.md`.

#### Régimen de `schemas/` y de `testdata/` (constitución, capa 3)

- **FR-085**: Toda modificación de ficheros ya existentes bajo `schemas/` o `testdata/` que este hito exige —`schemas/normas.yaml.json`, que gana el campo `vertebral` (FR-067); `schemas/eval.yaml.json`, que gana la variante de territorio y la regla del esperado verificable (FR-084); y cualquier fichero de grabaciones ya versionado que FR-072 reutilice— DEBE ir en una tarea `[datos]` con pausa humana, separada del bucle de implementación y sin mezclar otro trabajo (constitución, capa 3). Los dos cambios de esquema DEBEN ser compatibles hacia atrás: ningún fichero de `data/` ni de `evals/` ya válido deja de serlo.
- **FR-086**: Los ficheros nuevos que este hito añade bajo `schemas/` —el esquema de entrada y salida del applet (FR-007), el de los ficheros de `data/territorio/` y el del fichero de jerarquía (FR-044, FR-067)— y bajo `testdata/` —el corpus de fuzz (FR-035), las grabaciones de las búsquedas del BOE (FR-073) y los guiones del e2e de la matriz territorial, que en este repositorio viven bajo el `testdata/` de su paquete (FR-090)— DEBEN crearse también en tareas `[datos]`, separadas del bucle de implementación, porque son material de verdad terreno y de contrato. La pausa humana de FR-085 alcanza además a todo fichero nuevo bajo `schemas/` o bajo el `testdata/` de la raíz; el material de test nuevo bajo el `testdata/` de un paquete no la exige, y lo revisa la revisión final del hito.

#### Controles, documentación y Definition of Done

- **FR-090**: DEBE haber un test e2e (`testscript`) con la matriz territorial completa: Leganés (cubierto), Tordesillas (datos nacionales completos, boletines autonómico y provincial declarados no cubiertos, nada inventado), un municipio de Navarra o del País Vasco (régimen foral marcado) y un nombre ambiguo.
- **FR-091**: El e2e DEBE comprobar el código de salida de cada caso y el contenido del sobre, incluido que en el caso no cubierto no aparece ningún boletín no configurado.
- **FR-092**: Toda salida del applet DEBE validarse contra `schemas/*.json` en test, y `make schema-check` DEBE fallar si `schemas/` no coincide con lo que `--describe` emite.
- **FR-093**: Los tests DEBEN correr offline; ninguno abre red.
- **FR-094**: `make ci` DEBE quedar en verde: formato, lint, tests con detector de carreras, tests de integración, vulnerabilidades y esquemas.
- **FR-095**: La cobertura DEBE mantenerse en `internal/core/**` ≥ 85 % y global ≥ 70 %.
- **FR-096**: Los errores DEBEN ser tipados y mapear a exit codes estables; ningún `panic` en rutas de usuario.
- **FR-097**: Las reglas de dependencia DEBEN seguir cumpliéndose: `internal/core/**` no importa adaptadores, solo `httpx` importa `net/http`, solo `cli` y `cmd/` llaman a `os.Exit`, solo `render` escribe en la salida estándar.
- **FR-098**: `CHANGELOG.md` (sección *Unreleased*) DEBE registrar el applet `territorio`, la skill `legal-core` y los datos nuevos de `data/territorio/`.
- **FR-099**: `README.md` y `CONTRIBUTING.md` DEBEN quedar alineados donde enumeran applets, skills, directorios de `data/` o pasos de trabajo afectados por este hito.

#### Aceptación

- **FR-100**: En una sesión de Claude Code, la pregunta «¿qué comunidad, provincia y boletines corresponden a mi ayuntamiento?» DEBE responderse para un municipio de la Comunidad de Madrid con su comunidad, su provincia y sus boletines.
- **FR-101**: La misma pregunta DEBE responderse para un municipio de cualquier otra comunidad, con la cobertura explícita de lo que no está configurado.
- **FR-102**: Mientras `.kitlegal/config.yaml` no exista (llega en H10), la skill DEBE obtener el municipio de la conversación y, si no se ha dicho cuál es, preguntarlo en lugar de suponerlo.

### Key Entities

- **Municipio**: unidad territorial con nombre oficial, código INE de cinco cifras y su dígito de control, provincia y comunidad autónoma. Su id natural es el código INE (en H7 será el id de nodo `Municipio`).
- **Ayuntamiento (órgano)**: la administración del municipio; su id natural es el código DIR3, derivado del número de inscripción del REL y verificado. Puede faltar: entonces se declara en `cobertura`.
- **Comunidad autónoma**: nivel con potestad legislativa; determina el régimen (común o foral) y, si está configurada, el boletín autonómico. Ceuta y Melilla son ciudades autónomas y se tratan como territorio sin configuración.
- **Boletín aplicable**: el diario oficial de un nivel (estatal, autonómico, provincial) que publica lo que obliga en ese municipio. En una comunidad uniprovincial, el autonómico hace también de provincial.
- **Régimen**: común o foral. Dato nacional, conocido para todo municipio.
- **Cobertura**: la declaración, aspecto por aspecto, de qué del territorio está configurado o verificado y qué no. Es lo que impide que una skill concluya «no existe».
- **Configuración de comunidad**: el fichero de `data/territorio/` que rellena lo que varía por territorio. Solo el de la Comunidad de Madrid está relleno.
- **Ley vertebral**: norma de la tabla de `refs/mapa-sistema-legal-skills.md` §1.4 que toda skill debe tener a mano, con su identificador `BOE-A-…` verificado y marcada `vertebral: true` en `data/normas.yaml`.
- **Jerarquía normativa**: niveles (UE, Estado, comunidad autónoma, provincia, municipio), clase de boletín de cada uno y reglas de interpretación, en un fichero de datos propio (FR-067).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Para el municipio de referencia de la Comunidad de Madrid, `territorio resolver` devuelve exit 0 y un `data` con los 8 elementos que pide la entrega (municipio, código INE, provincia, comunidad, DIR3, régimen, boletines y `cobertura`); resolver por nombre y por código INE da el mismo `data` y la misma huella; y con `--offline` la salida es byte a byte la misma.
- **SC-002**: Para un municipio de una comunidad sin configuración, exit es 0, los datos nacionales están completos, `cobertura` declara no configurados el boletín autonómico y el provincial, y una búsqueda sobre la salida del nombre y del código de cualquier boletín no configurado no devuelve ninguna coincidencia.
- **SC-003**: Un municipio de Navarra y uno del País Vasco salen con régimen foral; un municipio de cada una de las demás comunidades y de las dos ciudades autónomas sale con régimen común; el 100 % de los municipios de la relación tiene régimen.
- **SC-004**: Un nombre compartido por N municipios devuelve exit 2 y lista los N candidatos con su código INE y su provincia; un nombre inexistente y un código inexistente devuelven exit 3; un código mal formado devuelve exit 2; y el applet no decide nunca 4, 5 ni 6 (solo el plazo agotado de `--timeout`, que pone el kernel, termina en 4).
- **SC-005**: La configuración de la Comunidad de Madrid es el único fichero de comunidad relleno, declara el BOCM como autonómico y provincial con su razón, y añadir otra comunidad no exige tocar ningún fichero fuera de `data/territorio/`.
- **SC-006**: Los analizadores de código INE y DIR3 aceptan y rechazan según formato y dígito de control, su fuzz corre sin `panic` con el corpus versionado, y la ida y vuelta de toda entrada aceptada es estable.
- **SC-007**: `data/territorio/` contiene el registro de municipios de la relación oficial del INE y la correspondencia INE→DIR3, los dos validados contra su esquema, con su procedencia y la fecha del fichero de origen, y ninguna prueba ni ejecución del binario abre una conexión hacia el INE, el REL o DIR3.
- **SC-008**: El registro de la verificación de la tarea `[datos]` cubre los 3 casos del control (municipio fusionado o renombrado, foral y con entidades locales menores) con el código derivado y el código real de cada uno; y todo municipio no verificado sale sin DIR3 y declarado en `cobertura`.
- **SC-009**: `skills/legal-core/SKILL.md` pasa la comprobación mecánica de skills (frontmatter, menos de 300 líneas, tabla de comandos y `references/` sin drift), su protocolo empieza por identificar el territorio, y las dos referencias (`leyes_vertebrales.md` y `jerarquia_normativa.md`) se regeneran desde `data/` con diff vacío.
- **SC-010**: `data/normas.yaml` contiene todas las leyes vertebrales de la tabla de `refs/`, cada identificador nuevo coincide con el de su búsqueda grabada del BOE en una comprobación sin red, y `references/normas.md` queda sin drift.
- **SC-011**: `evals/legal-core/` incluye al menos una eval de municipio cubierto, una de no cubierto y una de no activación; el conjunto se lee sin ficheros mal formados; y las evals están commiteadas antes que el código de la skill (su verde lo mide SC-015).
- **SC-012**: `make ci` queda en verde sin red, con la cobertura de `internal/core/**` ≥ 85 % y global ≥ 70 %, el e2e de la matriz territorial en verde y `make schema-check` sin drift.
- **SC-013**: En la ejecución de aceptación, la pregunta de la aceptación se responde para un municipio de la Comunidad de Madrid y para uno de otra comunidad; en el segundo caso la respuesta dice explícitamente qué no está configurado y no nombra ningún boletín que el applet no haya devuelto.
- **SC-014**: `CHANGELOG.md` (*Unreleased*), `README.md` y `CONTRIBUTING.md` registran el applet, la skill y los datos nuevos; `docs/SOURCES.md` sigue siendo cierto respecto de los ficheros generados y `scripts/verify-sources.sh` no gana ningún caso para las fuentes congeladas.
- **SC-015**: En una misma ejecución del job de evals (ADR 0016), **todas** las evals de `evals/legal-core/` pasan —las de activación positiva (municipio cubierto y no cubierto, FR-080 y FR-081) y la de no activación (FR-082)—, midiendo «pasa» por serie y umbral como H5: una eval pasa cuando al menos 2 de sus 3 sesiones pasan, y el informe lleva el commit evaluado y el id del modelo; en esa ejecución ninguna petición llegó a la red de una fuente. En la misma ejecución, el conjunto de `evals/boe-legislacion/` sigue pasando con su regla de H5 y H5.1, sin que sus ficheros de eval se hayan modificado (FR-074, FR-084). El commit del informe es de la rama del hito, y la cabeza que se fusiona solo difiere de él en ficheros bajo `specs/008-h6-territorio-skill-legal/`.

## Fuera de alcance

Lo que no está en el hito, en `CLAUDE.md`, en `refs/` ni en la constitución no se implementa. En concreto:

- **Festivos** de cualquier nivel y cualquier cálculo de plazos: son de H9, incluida la carga de los festivos locales de la Comunidad de Madrid que ADR 0017 congela en `data/festivos/`.
- **Operaciones de grafo**: `territorio` no emite `Municipio` ni `Organo` en este hito; las incorpora H7 (ADR 0014), que es también quien crea `internal/graph` y `world.db`.
- **`.kitlegal/config.yaml`** con el municipio de la persona usuaria, `--asunto` y el grafo del asunto: son de H10.
- **Otros verbos de `territorio`** (listar, buscar, órganos, competencia…): el hito entrega `resolver`.
- **Otros territorios configurados**: ninguna comunidad salvo Madrid se rellena; BOCYL, BOP de provincias de comunidades multiprovinciales, boletines forales y demás llegan desde el backlog («territorios»), bajo demanda.
- **`data/boletines/`** y el motor genérico de boletines: llega con H14 y siguientes; aquí los boletines son un dato de la configuración de comunidad, no un adaptador.
- **`data/organos.yaml`** y cualquier órgano distinto del ayuntamiento (diputación, órganos de control externo, plataformas autonómicas).
- **Consultar DIR3, el REL o el INE por red**, en ejecución o en grabación, y cualquier intento de sortear un WAF, un CAPTCHA, un certificado o la Red SARA (ADR 0017, decisión 4; constitución, principio I).
- **Actualización automática** de los ficheros congelados: envejecen con el repositorio y se refrescan repitiendo la tarea `[datos]`.
- **Identificadores distintos del código INE y del DIR3** en `internal/core/ids` (ELI, ECLI, CELEX, NIF): entran con sus hitos.
- **`internal/core/competencia`**, el reparto competencial ejecutable y `competencia quien-regula`: `legal-core` v0 trae la jerarquía como referencia, no como herramienta.
- **`legal-core` v1** (`recursos_y_plazos.md`, `reglas_de_cita.md`, uso de `plazos`): es H9.
- **Cambiar el protocolo, las referencias o las evals de `boe-legislacion`** más allá de regenerar `references/normas.md` por las normas nuevas y de mantener sus reglas del conjunto en verde.
- **Packs por vertical**, campo `vertical:` en `data/normas.yaml` y cualquier especialización vertical de `legal-core` (ADR 0012; ROADMAP §5).
- **Geometría, callejero, entidades locales menores como entidad propia, pedanías, mancomunidades y padrón**: el hito resuelve municipio, no territorio físico ni submunicipal.
- **Un tercer estado en `cobertura`** que distinga, por fila, el DIR3 comprobado individualmente del derivado por la regla: la evidencia de la muestra vive en el registro de verificación (FR-047) y la procedencia de cada fila en su `source`.
- **Que `legal-core` invoque el applet `boe`** o lleve sus verbos en su tabla de comandos, y que sus evals midan citas de norma y bloque del BOE: es de `boe-legislacion` (FR-069).
- **Leer `data/territorio/` del sistema de ficheros en ejecución** o mediante una ruta configurable: los datos van empaquetados (FR-056).
- **Un formato o juicio de evals propio de `legal-core`**, separado del común: se extiende el común (FR-084).
- **Refactorizar `internal/cli`** o extraerlo a librería (ROADMAP §5; se revisa al cerrar H9).
- **Un ADR nuevo**, salvo que la implementación cambie una decisión de arquitectura ya tomada: ADR 0017 ya decide el origen de los datos y ADR 0006 la procedencia del sobre de un applet que no consulta fuentes.

## Assumptions

- **Procedencia del sobre**: `territorio` encaja en la fila «applet calculado» de ADR 0006 —no consulta ninguna fuente en ejecución—, así que su `fuente` y su `url` van en el espacio reservado `kitlegal.` / `kitlegal:`, y los identificadores de fuente de `docs/SOURCES.md` (`ine.municipios`, `ine.codigos-territoriales`, `mpt.rel`) viajan dentro de `data`, en el `source` de cada dato, como dice la propia cabecera de la tabla de fuentes congeladas y ADR 0017.
- **Régimen como dato nacional**: el control exige régimen foral marcado en Navarra y el País Vasco, comunidades sin configuración, de modo que el régimen no puede vivir en la configuración por comunidad que solo Madrid rellena.
- **La relación del INE como registro de municipios**: es la única fuente que `docs/SOURCES.md` documenta para municipios, con dígito de control oficial y fecha de referencia a 1 de enero de 2026.
- **La derivación del DIR3** (`L` + número de inscripción del REL, que incorpora el código INE y su dígito de control) es una hipótesis hasta que la tarea `[datos]` la verifique; el spec no la da por cierta en ningún requisito (FR-046 a FR-048).
- **La verificación contra «DIR3 real» es manual**: ADR 0017 constata que ningún volcado DIR3 es descargable de forma automatizada, así que la muestra la comprueba una persona dentro de la tarea `[datos]` y deja registrada la evidencia (FR-047).
- **La tabla de leyes vertebrales** es la de `refs/mapa-sistema-legal-skills.md` §1.4, cuyo propio texto advierte de que algunos identificadores están escritos de memoria y hay que verificarlos; H5 dejó dicho que «la tabla completa se cierra en H6».
- **El municipio de referencia** es Leganés y el no cubierto es Tordesillas, como fija el control; son fixtures y casos de prueba, no casos especiales del producto.
- **Dependencias**: el kernel de H1 (flags globales, exit codes, sobre), `internal/httpx` de H2 y `internal/cache` de H3 (que este applet no necesita, porque no consulta red), el applet `boe` de H4 —para verificar los identificadores con `boe buscar` sobre grabaciones— y el andamiaje de skills y evals de H5 con las enmiendas de ADR 0016 y de H5.1.
- **Ejecución de aceptación**: se mide como en H5 y H5.1, con el job de evals que dispara la propuesta de cambio (ADR 0016), sobre el modelo del uso real y con sus repeticiones; la forma exacta de contar el verde está en SC-015, y SC-013 es la comprobación de la respuesta en esa misma ejecución.
- **Documentación del repositorio**: alinear `README.md` y `CONTRIBUTING.md` cuando un hito añade un applet, una skill o un directorio de `data/` es la práctica fijada en el repositorio (PR #35) y la lectura habitual de la Definition of Done §1.6; por eso FR-099 está en el spec y no en *Fuera de alcance*.
