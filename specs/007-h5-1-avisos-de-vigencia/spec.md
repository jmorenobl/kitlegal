# Feature Specification: H5.1 · Avisos de vigencia en las evals

**Feature Branch**: `007-h5-1-avisos-de-vigencia`

**Created**: 2026-09-16

**Status**: Draft

**Input**: Sección `#### H5.1 · Avisos de vigencia en las evals` de `docs/ROADMAP.md` (modo desatendido).

## Resumen

H5 dejó la skill `boe-legislacion` y el andamiaje que la mide: un formato común de eval que exige comandos ejecutados y citas encontradas, un `Juzgar` mecánico que reparte lo esperado en encontrado y ausente, y un informe que lo publica. El binario ya emite, en el sobre de `articulo`, `articulos` y `metadatos`, los avisos de vigencia de cada norma (`consolidacion-no-finalizada`, `derogada`, `vigencia-agotada`), y `SKILL.md` pide trasladarlos (regla 3). Pero ninguna eval comprueba que la respuesta los traslade: las 17 evals de `boe-legislacion` leen normas vivas, y el formato no sabría exigirlo (bitácora `docs/USO.md`, entrada del 2026-09-16 «Ninguna eval comprueba qué hace la skill ante una norma derogada»).

H5.1 no entrega skill nueva: **protege la que existe** (constitución, principio VIII) con cuatro piezas:

1. **Una forma fija para cada aviso en la respuesta.** `SKILL.md` fija cómo se traslada cada aviso de vigencia —`⚠`, la etiqueta del aviso tal como la da el binario y dos puntos, con la frase del binario o una explicación detrás—, igual que ya fija la forma de la cita. La etiqueta de cada código la exporta el binario, que pasa a ser la única fuente de verdad de la forma, sin cambiar lo que el applet emite.
2. **`avisos` en el formato común de eval**: la lista de códigos de aviso que la respuesta tiene que llevar, válida solo en evals que activan la skill.
3. **Juicio mecánico por la forma fija**, sin modelo ni lectura de la redacción libre: el aviso está en la respuesta si la respuesta lleva su forma fija, con tolerancia solo a lo que no cambia qué aviso es. Un aviso ausente impide que la eval pase, y el informe publica encontrados, ausentes y motivos.
4. **Una eval sobre una norma derogada**, la Ley 30/1992 (`BOE-A-1992-26318`), informativa, que exige sus dos avisos.

## Clarifications

### Session 2026-09-16

- Q: ¿Qué debe publicar el informe para que «su tasa dice en cuántas de sus tres sesiones la skill los trasladó» (FR-071, SC-007) quede cumplido para la eval de la norma derogada? → A: Basta con lo que el informe ya publica, sin ningún recuento nuevo. La tasa de la eval de la norma derogada es la tasa de su serie, la de ADR 0016: sesiones leídas y cuántas pasan, sobre las tres repeticiones del modelo que decide. Con FR-031 una sesión no pasa si le falta algún aviso esperado, así que esa tasa ya mide que se trasladaron los avisos. Cada una de las tres sesiones publica además `avisos_encontrados`, `avisos_ausentes` y un motivo por cada ausente (FR-040 a FR-042). La aceptación de FR-071 y SC-007 se lee así: la eval declara sus dos avisos esperados, `derogada` y `vigencia-agotada`; su fila de tasas da la tasa sobre 3 sesiones; y el reparto de avisos de cada sesión dice en cuáles se trasladaron, aunque esa sesión haya fallado por comandos o citas. (auto: criterio a; fuente: `docs/ROADMAP.md` §4 H5.1, Alcance, Entrega y Aceptación; ADR 0016, Decisión 2; constitución, Criterio de decisión autónoma punto 2; spec.md, Assumptions)
- Q: ¿Qué tiene que encontrar en `skills/boe-legislacion/SKILL.md` la comprobación mecánica de FR-014 para dar por presente la etiqueta de un código de aviso? → A: La comprobación de FR-014 da por presente un código cuando `SKILL.md` lleva, en cualquier parte, la forma fija completa de ese aviso: `⚠`, la etiqueta que exporta `internal/source/boe` y dos puntos. La reconoce con la misma función que usa `Juzgar`, con las mismas tolerancias (variante de presentación del emoji, énfasis de Markdown, espacios y mayúsculas) y exacta en la etiqueta. Si falta la forma de un código, falla nombrando ese código: esto obliga a que `SKILL.md` escriba literalmente la forma de cada uno de los tres avisos (p. ej. `⚠ NORMA DEROGADA:`), no solo la etiqueta suelta dentro de una descripción genérica. (auto: criterio c; fuente: `docs/ROADMAP.md` §4 H5.1, Controles, Alcance y Decisión del mecanismo; constitución, Gates capa 1 y Criterio de decisión autónoma punto 1; spec.md FR-011, FR-014, US4)

## Criterios del hito, literales

Transcripción literal de `docs/ROADMAP.md` §4, H5.1, para que la trazabilidad del spec se pueda comprobar sin volver al roadmap.

- **Objetivo**: «que las evals midan lo que hoy no comprueba nadie: que la skill **traslada a su respuesta** los avisos de vigencia que emite el binario. No entrega skill nueva; protege la que ya existe (constitución, principio VIII). Entra desde la bitácora `docs/USO.md`, entrada del 2026-09-16 «Ninguna eval comprueba qué hace la skill ante una norma derogada» (ADR 0013, §6). Va antes de H6 para que el formato de eval deje de crecer justo cuando un hito empieza a apoyarse en él.»
- **Entrega**: «`SKILL.md` de `boe-legislacion` fija la forma con la que la respuesta traslada cada aviso de vigencia, como ya fija la de la cita; el formato común de eval gana `avisos`, la lista de códigos de aviso que la respuesta tiene que llevar; `Juzgar` los reparte en encontrados y ausentes como hace con las citas, un aviso ausente impide que la eval pase, y el informe los publica; y `evals/boe-legislacion/` gana una eval sobre una norma derogada que los exige.»
- **Decisión del mecanismo** [2026-09-16, Jorge]: «el aviso se compara **por su forma fija, como la cita por su identificador**, y no leyendo la redacción libre de la respuesta. […] Un juicio semántico (modelo o similitud) queda descartado: lo prohíbe este alcance y la constitución (capa 1), y la similitud no separa «ha sido derogada» de «no ha sido derogada». Con la forma fija, lo que se mide es una regla escrita de la skill, la negación deja de importar —una respuesta que afirma lo contrario no lleva la forma— y un aviso nuevo del binario es una etiqueta más, sin reglas nuevas; además, el aviso llega siempre visible y con la misma forma a quien usa la skill.»
- **Alcance**: «`skills/boe-legislacion/SKILL.md` (regla 3 y paso 5, y donde explica cómo se cita): cada aviso de vigencia del sobre se traslada con su forma fija, `⚠` seguido de la etiqueta del aviso tal como la da el binario (`NORMA DEROGADA`, `VIGENCIA AGOTADA`, `TEXTO POSIBLEMENTE DESACTUALIZADO`) y dos puntos, con la frase del binario o una explicación detrás; sin nombrar evals, el job ni modelos (FR-077 de H5), con la región generada intacta y menos de 300 líneas. `internal/source/boe`: exportar la etiqueta de cada código de aviso, única fuente de verdad de la forma, sin cambiar códigos, frases, condiciones ni salida del applet. `schemas/eval.yaml.json` e `internal/evals/formato.go` (campo `avisos`, valores del enumerado de `internal/source/boe.CodigosDeAviso()`, permitido solo con `activa: true`). La comparación en `internal/evals/juzgar.go`, **mecánica y sin juicio de ningún modelo**: un aviso está en la respuesta si la respuesta lleva su forma fija, con tolerancia solo a lo que no cambia qué aviso es (la variante de presentación del emoji, el énfasis de Markdown, los espacios y las mayúsculas) y exacta en la etiqueta, como la extracción de la cita tras T046 de H5 es tolerante con lo que rodea al identificador y exacta en él; el plan fija esa tolerancia al detalle. Así **distingue una respuesta que traslada el aviso de otra que afirma lo contrario**: la que dice que la norma sigue en vigor, o traslada el aviso con otra redacción, no lleva la forma y el aviso queda ausente. `avisos_encontrados`, `avisos_ausentes` y sus motivos en `internal/evals/informe.go`. La norma de la eval es `BOE-A-1992-26318` (Ley 30/1992): sus metadatos, ya grabados en H4, dan `estatus_derogacion` y `vigencia_agotada` afirmativos —los dos avisos `derogada` y `vigencia-agotada`—, su bloque `a42` también está grabado y su `buscar` lo resuelve la búsqueda grabada en H4 `procedimiento administrativo común` (`internal/source/boe/testdata/golden/buscar-procedimiento-administrativo-comun.json`); falta grabar su `indice`, que FR-074 exige del índice y los metadatos de toda norma de una eval. Eso entra en `testdata/evals/grabaciones.json` y lo graba una persona en una tarea `[datos]` con pausa, nunca dentro de un job. La norma entra en `data/normas.yaml`, que obliga a regenerar `references/normas.md` con `make skills-sync`. La eval nace `informativa: true` (ADR 0016): así no toca la regla «positivas», que pide exactamente 10 que deciden, ni «materias distintas».»
- **Fuera de alcance**: «promover la eval a decisoria, que obligaría a enmendar FR-062 y su regla, y que se decide con los datos de varias ejecuciones; marcar la derogación en `data/normas.yaml` o en `references/`, porque el aviso lo emite el binario en el momento de consultar y no la tabla; extender los avisos a otras skills o fuentes; juzgar la redacción libre de la respuesta con listas de expresiones, reglas de atribución a normas o cualquier forma de similitud (ver *Decisión del mecanismo*), incluido el caso de una respuesta que lleva la forma fija y además dice lo contrario, que queda como limitación declarada y visible en el informe, que publica la respuesta; cambiar en `SKILL.md` algo distinto de la forma de trasladar los avisos; y tocar `internal/source/boe` más allá de exportar las etiquetas de aviso.»
- **Controles**: «`make ci` en verde, con el drift de lo generado y las reglas del conjunto de evals; tests de formato (un `avisos` con un código desconocido o en una eval de no activación es un fichero mal formado); de `Juzgar` (aviso esperado con su forma fija, también con las variantes toleradas; ausente; trasladado con otra redacción, que queda ausente; y una respuesta que niega el aviso, que no puede pasar); del informe; una comprobación mecánica en `make ci` de que los códigos del esquema coinciden con `CodigosDeAviso()` y de que `SKILL.md` lleva la etiqueta de cada uno, que falla nombrando el código; y la ejecución del job de evals que abre la propuesta de cambio.»
- **Aceptación**: «en el informe de esa ejecución, la eval de la norma derogada declara sus dos avisos esperados y su tasa dice en cuántas de sus tres sesiones la skill los trasladó; el veredicto sigue aprobado, porque el hito cambia `SKILL.md` (Definition of Done §1.10); y `red` sigue vacío.»

Trazabilidad resumida (el detalle, en cada requisito y criterio):

| Criterio literal | Dónde se cumple |
|---|---|
| `SKILL.md` fija la forma con la que la respuesta traslada cada aviso (regla 3, paso 5, «Cómo se cita») | FR-001 a FR-003, US1 escenarios 1 y 2, SC-001 |
| cambiar en `SKILL.md` solo la forma de trasladar los avisos (*Fuera de alcance* del hito) | FR-004, US1 escenario 4, SC-001 |
| sin nombrar evals, el job ni modelos; región generada intacta; menos de 300 líneas | FR-005, SC-001 |
| `internal/source/boe` exporta la etiqueta de cada código, única fuente de verdad, sin cambiar códigos, frases, condiciones ni salida | FR-010 a FR-012, US4 escenarios 1, 4 y 5, SC-005 |
| el formato común de eval gana `avisos`, con valores del enumerado de `CodigosDeAviso()`, solo con `activa: true` | FR-020 a FR-023, US2 escenarios 1, 2, 9 y 10, SC-002 |
| `Juzgar` reparte los avisos en encontrados y ausentes; un aviso ausente impide que la eval pase | FR-030, FR-031, FR-034, US2 |
| comparación mecánica por la forma fija, sin modelo, tolerante solo con emoji, énfasis, espacios y mayúsculas, exacta en la etiqueta | FR-032, FR-033, US2, SC-003 |
| distingue la respuesta que traslada el aviso de la que afirma lo contrario o usa otra redacción | FR-033, US2 escenarios 4 y 5, SC-003 |
| el informe publica `avisos_encontrados`, `avisos_ausentes` y sus motivos | FR-040 a FR-042, US3, SC-004 |
| eval sobre `BOE-A-1992-26318` con los avisos `derogada` y `vigencia-agotada`, `informativa: true` | FR-050, FR-051, FR-053, US5 escenarios 1 y 2, SC-006 |
| la eval se escribe antes que el cambio de `SKILL.md` (Definition of Done §1.10) | FR-052, US5 escenario 5, SC-009 |
| grabar el `indice` de la norma en `testdata/evals/grabaciones.json` en una tarea `[datos]` con pausa | FR-054, FR-055, FR-057, US5 escenario 3, SC-006, SC-009 |
| la norma entra en `data/normas.yaml` y se regenera `references/normas.md` con `make skills-sync` | FR-056, SC-006 |
| **Control**: `make ci` en verde, con el drift de lo generado y las reglas del conjunto de evals | FR-060, SC-006 |
| **Control**: tests de formato (código desconocido; `avisos` en eval de no activación) | FR-021, FR-022, US2 escenarios 1 y 2, SC-002 |
| **Control**: tests de `Juzgar` (forma fija, variantes toleradas, ausente, otra redacción, negación) | FR-035, US2 escenarios 3 a 6, SC-003 |
| **Control**: tests del informe | FR-043, US3, SC-004 |
| **Control**: comprobación mecánica en `make ci` esquema ↔ `CodigosDeAviso()` y `SKILL.md` ↔ forma fija completa (`⚠`, etiqueta y dos puntos) de cada código, que falla nombrando el código (Clarifications, Q2) | FR-013, FR-014, US4 escenarios 2 y 3, SC-005 |
| **Control**: la ejecución del job de evals que abre la propuesta de cambio, con su informe y su commit registrados en el directorio del hito | FR-070, SC-007 |
| documentación: *Unreleased* de `CHANGELOG.md` y el formato de eval en `README.md` y `CONTRIBUTING.md` (Definition of Done §1.6) | FR-061, SC-008 |
| **Aceptación**: la eval de la norma derogada declara sus dos avisos esperados y su tasa dice en cuántas de sus tres sesiones la skill los trasladó | FR-071, SC-007 |
| **Aceptación**: el veredicto sigue aprobado | FR-072, SC-007 |
| **Aceptación**: `red` sigue vacío | FR-073, SC-007 |

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Quien consulta una norma derogada ve el aviso, siempre con la misma forma (Priority: P1)

Una persona pregunta a la skill por un artículo de una norma que el BOE da por derogada o con la vigencia agotada. El sobre que devuelve el binario trae los avisos de vigencia; la respuesta de la skill los traslada con una forma fija y visible —`⚠`, la etiqueta del aviso y dos puntos, seguidos de la frase del binario o de una explicación—, de modo que la persona no confunde el texto de una norma derogada con derecho vigente, y cualquier aviso llega siempre con la misma forma.

**Why this priority**: es lo que protege a quien usa la skill y lo único que la eval puede medir: sin una regla escrita en `SKILL.md` no hay forma fija que comparar.

**Independent Test**: se comprueba leyendo `SKILL.md` —regla 3, paso 5 y «Cómo se cita» describen la forma fija y escriben la forma fija completa de cada código—, con su diff contra `main`, y con la comprobación mecánica de `make ci` que exige en `SKILL.md` la forma fija completa de cada código (FR-014).

**Acceptance Scenarios**:

1. **Given** `SKILL.md` de `boe-legislacion`, **When** se lee la regla 3, **Then** dice que cada aviso de vigencia del sobre se traslada con su forma fija: `⚠`, la etiqueta del aviso tal como la da el binario y dos puntos, con la frase del binario o una explicación detrás.
2. **Given** `SKILL.md`, **When** se leen el paso 5 y la sección que explica cómo se cita, **Then** los dos describen la misma forma fija, y la explicación de cómo se cita escribe la forma fija completa de cada uno de los tres avisos —`⚠ NORMA DEROGADA:`, `⚠ VIGENCIA AGOTADA:` y `⚠ TEXTO POSIBLEMENTE DESACTUALIZADO:`—, no solo la etiqueta suelta, y al menos un ejemplo de aviso trasladado.
3. **Given** `SKILL.md` tras el cambio, **When** se ejecuta `make ci`, **Then** la comprobación de skills pasa: frontmatter válido, menos de 300 líneas, tabla de comandos generada sin drift, y el texto no nombra las evals, el job ni ningún modelo.
4. **Given** el diff de `SKILL.md` entre `main` y la rama del hito, **When** se revisa cada cambio, **Then** todos caen en la regla 3, el paso 5 o la explicación de cómo se cita, y dentro de ellos solo tocan la forma de trasladar los avisos.

---

### User Story 2 - Quien mantiene las evals exige avisos y el juicio los compara por su forma fija (Priority: P1)

Quien escribe una eval declara, en una eval que activa la skill, los códigos de aviso que la respuesta tiene que llevar. Al juzgar una sesión, cada aviso esperado queda encontrado si la respuesta lleva su forma fija —aunque varíe la presentación del emoji, el énfasis de Markdown, los espacios o las mayúsculas— y ausente en cualquier otro caso, también cuando la respuesta traslada el aviso con otra redacción o afirma lo contrario. Un aviso ausente impide que la eval pase.

**Why this priority**: es el mecanismo que hace medible la regla de US1, y lo que el roadmap decidió (*Decisión del mecanismo*).

**Independent Test**: tests del formato y de `Juzgar` sobre ficheros de eval y respuestas escritas en el test, sin modelo ni red.

**Acceptance Scenarios**:

1. **Given** una eval con `activa: true` y `avisos` con un código que no está entre los de `CodigosDeAviso()`, **When** se lee, **Then** es un fichero mal formado y el error nombra el fichero.
2. **Given** una eval con `activa: false` y `avisos`, **When** se lee, **Then** es un fichero mal formado y el error nombra el fichero.
3. **Given** una eval que espera `derogada` y una respuesta que lleva `⚠ NORMA DEROGADA:` seguido de la frase del binario, **When** se juzga, **Then** `derogada` queda entre los avisos encontrados y no entre los ausentes.
4. **Given** la misma eval y respuestas que llevan la forma fija con cada variante tolerada —el emoji con su selector de presentación, la forma con énfasis de Markdown, espacios distintos alrededor de sus partes, la etiqueta en otra combinación de mayúsculas y minúsculas—, **When** se juzga cada una, **Then** `derogada` queda encontrado en todas.
5. **Given** la misma eval y una respuesta que dice que la norma está derogada con otra redacción, sin la forma fija (p. ej. «Esta ley fue derogada.»), **When** se juzga, **Then** `derogada` queda ausente y la eval no pasa.
6. **Given** la misma eval y una respuesta que niega el aviso (p. ej. «La Ley 30/1992 sigue en vigor.»), **When** se juzga, **Then** `derogada` queda ausente y la eval no pasa aunque se hayan ejecutado todos los comandos y encontrado todas las citas.
7. **Given** una respuesta sin ninguna forma fija, **When** se juzga una eval que espera `derogada` y `vigencia-agotada`, **Then** los dos quedan ausentes, en el orden de la eval, y la eval no pasa.
8. **Given** una eval sin `avisos`, **When** se juzga cualquier sesión, **Then** encontrados y ausentes quedan vacíos y el resultado de la eval es el mismo que antes de H5.1.
9. **Given** una eval con `activa: true` y `avisos` vacío, y otra con un código repetido en `avisos`, **When** se leen, **Then** cada una se acepta o se rechaza exactamente como se acepta o se rechaza la misma eval con `citas` vacío o con una cita repetida.
10. **Given** una eval con `activa: true` y `avisos` con `derogada` y `vigencia-agotada`, **When** se lee, **Then** es válida y sus avisos esperados son esos dos códigos, en el orden del fichero.

---

### User Story 3 - El informe publica qué avisos llevó cada sesión y por qué no pasa (Priority: P2)

Quien lee el informe de una ejecución del job ve, por cada sesión, qué avisos esperados llevó la respuesta y cuáles no, y un motivo por cada aviso ausente; y ve la respuesta publicada, que es donde queda visible la limitación declarada de una respuesta que lleva la forma fija y además dice lo contrario.

**Why this priority**: sin publicarlo, un rojo por aviso ausente no se distingue de uno por cita ausente, y la aceptación del hito se lee en el informe.

**Independent Test**: tests del informe con sesiones y evals escritas en el test.

**Acceptance Scenarios**:

1. **Given** una sesión juzgada con una eval que espera dos avisos y una respuesta que lleva uno, **When** se escribe el informe, **Then** su resultado estructurado lleva `avisos_encontrados` con uno y `avisos_ausentes` con el otro, y la forma legible del informe los muestra junto a sus citas y sigue publicando la respuesta de la sesión.
2. **Given** esa misma sesión, **When** se escribe el informe, **Then** sus motivos incluyen uno por el aviso ausente, que nombra su código, detrás de los motivos de las citas ausentes.
3. **Given** una sesión juzgada con una eval sin `avisos`, **When** se escribe el informe, **Then** `avisos_encontrados` y `avisos_ausentes` aparecen como listas vacías.

---

### User Story 4 - La forma de cada aviso tiene una sola fuente de verdad, vigilada en `make ci` (Priority: P2)

Quien añade un aviso nuevo al binario no tiene que escribir reglas nuevas de juicio: añade su código y su etiqueta en un único sitio. `make ci` falla, nombrando el código, si el esquema de eval no admite exactamente los códigos del binario o si `SKILL.md` no lleva la forma fija completa (`⚠`, la etiqueta y dos puntos) de alguno.

**Why this priority**: es lo que hace que el mecanismo escale (*Decisión del mecanismo*) y que la regla escrita de la skill no se desincronice del binario.

**Independent Test**: la comprobación mecánica se prueba con los casos de los escenarios 2 y 3: un esquema con un código de menos y otro con un código de más, y un `SKILL.md` al que le falta la forma de un código porque le falta la etiqueta, el `⚠` o los dos puntos. Que la etiqueta tiene una sola fuente se prueba con el escenario 5.

**Acceptance Scenarios**:

1. **Given** el binario tras H5.1, **When** se consulta la etiqueta de cada código de aviso, **Then** hay exactamente una por cada código de `CodigosDeAviso()`: `consolidacion-no-finalizada` → `TEXTO POSIBLEMENTE DESACTUALIZADO`, `derogada` → `NORMA DEROGADA`, `vigencia-agotada` → `VIGENCIA AGOTADA`.
2. **Given** un esquema de eval cuyo enumerado de `avisos` no coincide con `CodigosDeAviso()` (le falta un código o le sobra uno), **When** corre la comprobación mecánica, **Then** falla nombrando el código que falta o sobra.
3. **Given** un `SKILL.md` al que le falta la forma fija completa de un código —falta la etiqueta, el `⚠` o los dos puntos—, **When** corre la comprobación mecánica, **Then** falla nombrando ese código.
4. **Given** las salidas de `articulo`, `articulos` y `metadatos` grabadas antes de H5.1, **When** se ejecutan los tests del applet y `make schema-check` tras exportar las etiquetas, **Then** los códigos, las frases, las condiciones y la salida del applet son idénticos y `schemas/norma.json` y `schemas/bloque.json` no cambian.
5. **Given** el código de `Juzgar` y el de la comprobación mecánica de FR-014, sin contar sus tests, **When** se buscan en él los literales `NORMA DEROGADA`, `VIGENCIA AGOTADA` y `TEXTO POSIBLEMENTE DESACTUALIZADO`, **Then** no aparece ninguno: los dos toman la etiqueta de cada código de lo que exporta `internal/source/boe`.

---

### User Story 5 - Una eval sobre una norma derogada, medida sin red y sin decidir el veredicto (Priority: P2)

`evals/boe-legislacion/` gana una eval sobre la Ley 30/1992 (`BOE-A-1992-26318`), derogada y con la vigencia agotada, que exige los avisos `derogada` y `vigencia-agotada`. Nace informativa: se ejecuta con el modelo que decide y su tasa se publica, pero no decide el veredicto, así que no altera las reglas «positivas» ni «materias distintas».

**Why this priority**: es la eval que la bitácora echaba en falta y la que da la aceptación del hito.

**Independent Test**: `make ci` valida el fichero, las reglas del conjunto y la comprobación sin red de lo grabado; la ejecución del job la mide.

**Acceptance Scenarios**:

1. **Given** la eval nueva, **When** se lee, **Then** es válida, activa la skill, es `informativa: true`, lleva sus comandos y citas esperados sobre `BOE-A-1992-26318` y `avisos` con `derogada` y `vigencia-agotada`.
2. **Given** `evals/boe-legislacion/` con la eval nueva, **When** corren las reglas del conjunto en `make ci`, **Then** se cumplen todas, incluidas «positivas» (exactamente 10 que deciden), «materias distintas», «informativas», «tamaño» y «normas conocidas».
3. **Given** la caché preparada desde lo grabado, **When** corre la comprobación sin red de las evals, **Then** el `indice` y los `metadatos` de `BOE-A-1992-26318`, su bloque citado y cada comando esperado de la eval terminan en código 0.
4. **Given** `data/normas.yaml` con la norma nueva, **When** se ejecuta `make skills-sync`, **Then** `references/normas.md` la incluye y el control de drift pasa sin marcar la derogación en la tabla.
5. **Given** la historia de la rama del hito, **When** se ordenan sus commits, **Then** el commit que añade la eval nueva precede al primer commit que cambia `SKILL.md`.

---

### Edge Cases

- **Respuesta que lleva la forma fija y además dice lo contrario** (p. ej. `⚠ NORMA DEROGADA:` seguido de «pero sigue en vigor»): el aviso queda encontrado. Es la limitación declarada del mecanismo; no se juzga (*Fuera de alcance*) y queda visible porque el informe publica la respuesta.
- **Respuesta con la forma fija de un aviso que la eval no espera**: no cambia el juicio; como una invocación adicional, lo no esperado no hace fallar la eval.
- **Etiqueta alterada dentro de la forma** (una palabra de más, de menos o distinta, p. ej. `⚠ NORMA PARCIALMENTE DEROGADA:`): no es la forma fija del aviso, que es exacta en la etiqueta, y el aviso queda ausente.
- **Etiqueta sin `⚠` o sin los dos puntos** (p. ej. «NORMA DEROGADA» en mitad de una frase): no es la forma fija; el aviso queda ausente.
- **La forma fija aparece varias veces o una sola vez para dos normas**: basta con que aparezca una vez; el juicio no atribuye avisos a normas (*Fuera de alcance*).
- **`avisos` vacío o con un código repetido**: se tratan con el mismo criterio que el formato ya aplica a `citas` (FR-022).
- **Eval informativa con un aviso ausente**: la sesión no pasa y cuenta en la tasa publicada de la eval, pero no hace fallar el veredicto (ADR 0016).
- **Sesión que no lee los metadatos ni ningún artículo de la norma derogada**: el sobre no le da avisos que trasladar; si la respuesta no lleva la forma fija, los avisos quedan ausentes, igual que cualquier otra respuesta sin la forma.
- **Invocación de la eval nueva fuera de lo grabado**: se trata como cualquier otra (FR-074 y FR-076 de H5): no llega a la red, el informe la nombra y por sí sola no hace fallar la eval.

## Requirements *(mandatory)*

### Functional Requirements

#### La forma fija en `SKILL.md`

- **FR-001**: La regla 3 de `skills/boe-legislacion/SKILL.md` MUST decir que cada aviso de vigencia del sobre se traslada a la respuesta con su forma fija: `⚠` seguido de la etiqueta del aviso tal como la da el binario y de dos puntos, con la frase del binario o una explicación detrás.
- **FR-002**: El paso 5 del protocolo MUST pedir trasladar cada aviso con esa misma forma fija, en lugar de la instrucción genérica actual.
- **FR-003**: La parte de `SKILL.md` que explica cómo se cita MUST describir la forma fija del aviso junto a la de la cita, escribiendo la forma fija completa de cada uno de los tres avisos —`⚠ NORMA DEROGADA:`, `⚠ VIGENCIA AGOTADA:` y `⚠ TEXTO POSIBLEMENTE DESACTUALIZADO:`—, no solo la etiqueta suelta (Clarifications, Q2), y al menos un ejemplo de aviso trasladado.
- **FR-004**: `SKILL.md` MUST NOT cambiar nada distinto de la forma de trasladar los avisos.
- **FR-005**: `SKILL.md` MUST seguir sin nombrar las evals, el job ni ningún modelo (FR-077 de H5), con la región de la tabla de comandos generada intacta, frontmatter válido y menos de 300 líneas.

#### Etiquetas de aviso en el binario

- **FR-010**: `internal/source/boe` MUST exportar la etiqueta de cada código de `CodigosDeAviso()`: `consolidacion-no-finalizada` → `TEXTO POSIBLEMENTE DESACTUALIZADO`, `derogada` → `NORMA DEROGADA`, `vigencia-agotada` → `VIGENCIA AGOTADA`.
- **FR-011**: Esa etiqueta exportada MUST ser la única fuente de verdad de la forma fija: la comparación de `Juzgar` y la comprobación mecánica de FR-014 MUST tomar la etiqueta de ahí y no de una copia.
- **FR-012**: Exportar las etiquetas MUST NOT cambiar los códigos, las frases, las condiciones de los avisos ni la salida del applet: los golden de `articulo`, `articulos` y `metadatos`, `schemas/norma.json`, `schemas/bloque.json` y `make schema-check` quedan sin cambios.
- **FR-013**: MUST existir una comprobación mecánica, sin red ni modelo y dentro de `make ci`, de que los códigos que admite `avisos` en `schemas/eval.yaml.json` coinciden exactamente con `CodigosDeAviso()`; si falta o sobra un código, MUST fallar nombrándolo.
- **FR-014**: MUST existir una comprobación mecánica, sin red ni modelo y dentro de `make ci`, de que `skills/boe-legislacion/SKILL.md` lleva, en cualquier parte, la forma fija completa (`⚠`, la etiqueta y dos puntos) de cada código de `CodigosDeAviso()`, reconocida con la misma función tolerante que usa `Juzgar` (FR-032) y exacta en la etiqueta; si falta la forma de algún código, MUST fallar nombrando ese código.

#### Formato común de eval

- **FR-020**: El formato común de eval (`schemas/eval.yaml.json` y su lectura en `internal/evals/formato.go`) MUST admitir el campo opcional `avisos`: la lista de códigos de aviso que la respuesta tiene que llevar.
- **FR-021**: Cada valor de `avisos` MUST ser uno de los códigos de `CodigosDeAviso()`; un código desconocido MUST hacer el fichero mal formado, con el error nombrando el fichero.
- **FR-022**: `avisos` MUST estar permitido solo con `activa: true`; en una eval de no activación MUST hacer el fichero mal formado, con el error nombrando el fichero. La lista vacía y los códigos repetidos se tratan con el mismo criterio que el formato ya aplica a `citas`.
- **FR-023**: Las evals existentes, que no llevan `avisos`, MUST seguir siendo válidas sin cambios.

#### Juicio de los avisos

- **FR-030**: `Juzgar` MUST repartir los avisos esperados de la eval, en su orden, entre encontrados en la respuesta y ausentes de ella, como hace con las citas.
- **FR-031**: Una eval MUST pasar solo si, además de lo que ya exige FR-072 de H5, ningún aviso esperado queda ausente.
- **FR-032**: La comparación MUST ser mecánica y sin juicio de ningún modelo: un aviso está en la respuesta si la respuesta lleva su forma fija (`⚠`, su etiqueta y dos puntos), con tolerancia solo a lo que no cambia qué aviso es —la variante de presentación del emoji, el énfasis de Markdown, los espacios y las mayúsculas— y exacta en la etiqueta. El plan fija esa tolerancia al detalle.
- **FR-033**: La comparación MUST NOT leer la redacción libre de la respuesta: ni listas de expresiones, ni reglas de atribución a normas, ni similitud. Una respuesta que afirma que la norma sigue en vigor, o que traslada el aviso con otra redacción, no lleva la forma y el aviso queda ausente.
- **FR-034**: El juicio de las evals sin `avisos` MUST NOT cambiar: para ellas, encontrados y ausentes quedan vacíos y el resultado es el mismo que antes de H5.1.
- **FR-035**: MUST haber tests de `Juzgar`, sin red ni modelo, para: el aviso esperado con su forma fija; la forma fija con cada una de las variantes toleradas; el aviso ausente; el aviso trasladado con otra redacción, que queda ausente; y una respuesta que niega el aviso, que no puede pasar.

#### Informe

- **FR-040**: El resultado de cada sesión en el informe MUST publicar `avisos_encontrados` y `avisos_ausentes`, cada uno en el orden de la eval y como lista vacía cuando no hay ninguno.
- **FR-041**: Cada aviso ausente MUST dar un motivo por el que la eval no pasa, que nombra su código, detrás de los motivos de las citas ausentes.
- **FR-042**: La forma legible del informe MUST mostrar los avisos encontrados y ausentes de cada sesión junto a sus citas, y seguir publicando la respuesta de la sesión.
- **FR-043**: MUST haber tests del informe que cubran FR-040 a FR-042.

#### La eval de la norma derogada

- **FR-050**: `evals/boe-legislacion/` MUST ganar una eval sobre la norma derogada `BOE-A-1992-26318` (Ley 30/1992), con `activa: true` e `informativa: true`, y con `avisos` que exige `derogada` y `vigencia-agotada`.
- **FR-051**: Como toda eval que activa la skill, la eval MUST llevar comandos esperados y citas esperadas, todos sobre `BOE-A-1992-26318` y dentro de lo grabado (FR-054): el bloque `a42` como bloque esperado y como cita esperada. La redacción literal de la pregunta la fija el plan.
- **FR-052**: La eval MUST escribirse y commitearse antes que el cambio de `SKILL.md` (Definition of Done §1.10; ritual §6).
- **FR-053**: Con la eval nueva, el conjunto de `evals/boe-legislacion/` MUST seguir cumpliendo todas sus reglas: «positivas» sigue contando exactamente 10 que deciden y «materias distintas» no la cuenta, porque es informativa (ADR 0016).
- **FR-054**: Lo grabado para las evals MUST cubrir, para `BOE-A-1992-26318`, lo que FR-074 de H5 exige de toda norma de una eval —su `indice`, sus `metadatos`, el bloque de cada cita esperada y lo que necesite cada comando esperado— y la búsqueda con la que se verifica su identificador (FR-056), reutilizando lo ya grabado en H4 (metadatos, `a42` y la búsqueda `procedimiento administrativo común`) y añadiendo lo que falta, su `indice`, a `testdata/evals/grabaciones.json`.
- **FR-055**: Esa grabación MUST hacerla una persona en una tarea `[datos]` con pausa, nunca dentro de un job ni del bucle de implementación; y toda modificación de ficheros existentes en `testdata/` y `schemas/` —`testdata/evals/grabaciones.json` y `schemas/eval.yaml.json`— MUST ir en una tarea `[datos]` con pausa (constitución, capa 3).
- **FR-056**: `BOE-A-1992-26318` MUST entrar en `data/normas.yaml`, validada contra su esquema y con su identificador verificado contra la búsqueda grabada como el resto, y `references/normas.md` MUST regenerarse con `make skills-sync`. Ni `data/normas.yaml` ni `references/` MUST marcar la derogación.
- **FR-057**: La comprobación sin red de las evals (FR-075 de H5) MUST cubrir la eval nueva: su `indice`, sus `metadatos`, su bloque citado y cada comando esperado terminan en código 0 sobre la caché preparada.

#### Controles y documentación

- **FR-060**: `make ci` MUST quedar en verde, con el drift de lo generado (`references/`, tabla de comandos, symlinks de `scripts/`), las reglas del conjunto de evals, la comprobación sin red de las evals y las comprobaciones de FR-013 y FR-014.
- **FR-061**: `CHANGELOG.md` (sección *Unreleased*) MUST registrar la forma fija de los avisos en la skill, el campo `avisos` del formato y la eval nueva (Definition of Done §1.6); y la documentación del formato común de eval en `README.md` y `CONTRIBUTING.md`, que H5 dejó con los campos del formato (FR-060 y FR-083 de H5), MUST incluir `avisos`.

#### Aceptación

- **FR-070**: La ejecución de aceptación es la del job de evals que abre la propuesta de cambio del hito, sobre un commit de la rama del hito del que la cabeza que se fusiona solo difiere en ficheros bajo `specs/007-h5-1-avisos-de-vigencia/`; cualquier cambio posterior fuera de ese directorio MUST obligar a repetirla. Su informe, con el commit sobre el que corrió, MUST quedar enlazado o registrado en ese directorio. La tarea que la lanza y la lee lleva `[plataforma]`.
- **FR-071**: En el informe de esa ejecución, la eval de la norma derogada MUST declarar sus dos avisos esperados, y su tasa —la de su serie sobre las tres repeticiones del modelo que decide (ADR 0016)— MUST decir en cuántas de sus tres sesiones pasó; el reparto de avisos de cada sesión (FR-040 a FR-042) MUST decir en cuáles se trasladó cada aviso esperado, sin que el informe publique ningún recuento adicional aparte de la tasa de la eval y el reparto por sesión que ya exigen FR-040 a FR-042.
- **FR-072**: El veredicto de esa ejecución MUST seguir aprobado, porque el hito cambia `SKILL.md` (Definition of Done §1.10).
- **FR-073**: `red` MUST seguir vacío en el informe de esa ejecución: ninguna petición llegó a la red de una fuente (FR-076 de H5).

### Key Entities

- **Aviso de vigencia**: lo que el binario emite en el sobre cuando una condición de los metadatos de una norma lo pide. Tiene un código estable (`consolidacion-no-finalizada`, `derogada`, `vigencia-agotada`), una etiqueta (`TEXTO POSIBLEMENTE DESACTUALIZADO`, `NORMA DEROGADA`, `VIGENCIA AGOTADA`) y una frase que empieza por su forma fija.
- **Forma fija de un aviso**: `⚠`, la etiqueta y dos puntos. Es lo que la skill escribe en su respuesta y lo que el juicio busca; lo que va detrás es libre.
- **Aviso esperado**: un código de aviso declarado en `avisos` de una eval que activa la skill; la respuesta de cada sesión tiene que llevar su forma fija.
- **Reparto de avisos**: por sesión, los avisos esperados encontrados y ausentes, en el orden de la eval, con un motivo por cada ausente.
- **Eval de la norma derogada**: eval informativa sobre `BOE-A-1992-26318` que exige `derogada` y `vigencia-agotada`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `SKILL.md` describe la forma fija en la regla 3, el paso 5 y la explicación de cómo se cita, lleva la forma fija completa (`⚠`, la etiqueta y dos puntos) de cada uno de los 3 avisos, tiene menos de 300 líneas y no nombra evals, job ni modelos; la comprobación de skills de `make ci` pasa; y el 100 % de las líneas que su diff contra `main` cambia están en la regla 3, el paso 5 o la explicación de cómo se cita y solo tocan la forma de trasladar los avisos.
- **SC-002**: el 100 % de los ficheros de eval con un código de aviso desconocido o con `avisos` en una eval de no activación se rechazan como mal formados nombrando el fichero; `avisos` vacío o con un código repetido se acepta o se rechaza igual que `citas` en el mismo caso; una eval que activa la skill con `avisos` válidos se lee con sus códigos en el orden del fichero; y las 17 evals existentes siguen siendo válidas sin cambios.
- **SC-003**: los tests de `Juzgar` cubren los 5 casos del control (forma fija, variantes toleradas, ausente, otra redacción, negación): en los dos primeros el aviso queda encontrado; en los tres últimos queda ausente y la eval no pasa. Una eval sin `avisos` da encontrados y ausentes vacíos y el mismo resultado que antes de H5.1.
- **SC-004**: los tests del informe muestran `avisos_encontrados` y `avisos_ausentes`, cada uno en el orden de la eval, y un motivo por aviso ausente que nombra su código, detrás de los de las citas ausentes; listas vacías en las evals sin `avisos`; y la forma legible con los avisos junto a las citas y la respuesta de la sesión.
- **SC-005**: la comprobación mecánica falla nombrando el código en cada uno de sus 5 casos —en el esquema, un código de menos y un código de más; en `SKILL.md`, la forma fija de un código incompleta porque le falta la etiqueta, el `⚠` o los dos puntos— y pasa con el repositorio tal como queda; el binario exporta exactamente 3 etiquetas, una por cada código de `CodigosDeAviso()`; el código de `Juzgar` y el de la comprobación, sin contar sus tests, no contienen ningún literal de las 3 etiquetas; y la salida del applet y `schemas/norma.json` y `schemas/bloque.json` quedan byte a byte iguales.
- **SC-006**: `make ci` en verde con 18 evals en `evals/boe-legislacion/`: exactamente 10 positivas que deciden, la eval nueva informativa, `references/normas.md` sin drift e incluyendo `BOE-A-1992-26318`, la comprobación sin red en código 0 para todo lo de la eval nueva, y las comprobaciones de FR-013 y FR-014 en verde.
- **SC-007**: la ejecución de aceptación es la del job que abre la propuesta de cambio, sobre un commit del que la cabeza que se fusiona solo difiere en ficheros bajo `specs/007-h5-1-avisos-de-vigencia/`; su informe, con ese commit, queda enlazado o registrado en ese directorio; en él, la eval de la norma derogada declara 2 avisos esperados y su tasa sobre 3 sesiones, y cada una de esas 3 sesiones publica su reparto de avisos; el veredicto es aprobado; y `red` está vacío.
- **SC-008**: la sección *Unreleased* de `CHANGELOG.md` registra la forma fija de los avisos en la skill, el campo `avisos` del formato y la eval nueva; y la documentación del formato común de eval en `README.md` y en `CONTRIBUTING.md` incluye `avisos`.
- **SC-009**: en la historia de la rama del hito, el commit que añade la eval nueva precede al primer commit que cambia `SKILL.md`; y la grabación del `indice` de `BOE-A-1992-26318` y las modificaciones de `testdata/evals/grabaciones.json` y `schemas/eval.yaml.json` van en tareas `[datos]` con pausa, ninguna dentro de un job.

## Fuera de alcance

Del hito, literal:

- Promover la eval a decisoria, que obligaría a enmendar FR-062 y su regla, y que se decide con los datos de varias ejecuciones.
- Marcar la derogación en `data/normas.yaml` o en `references/`, porque el aviso lo emite el binario en el momento de consultar y no la tabla.
- Extender los avisos a otras skills o fuentes.
- Juzgar la redacción libre de la respuesta con listas de expresiones, reglas de atribución a normas o cualquier forma de similitud (ver *Decisión del mecanismo*), incluido el caso de una respuesta que lleva la forma fija y además dice lo contrario, que queda como limitación declarada y visible en el informe, que publica la respuesta.
- Cambiar en `SKILL.md` algo distinto de la forma de trasladar los avisos.
- Tocar `internal/source/boe` más allá de exportar las etiquetas de aviso.

No especificado en el hito, `CLAUDE.md`, `refs/` ni la constitución, y por tanto no se implementa:

- Juzgar el contenido que sigue a los dos puntos de la forma fija, o exigir la frase literal del binario.
- Hacer fallar una eval por la forma fija de un aviso que la eval no espera.
- Atribuir cada aviso a una norma concreta cuando la respuesta trata varias.
- Una eval para `consolidacion-no-finalizada` o para otra norma derogada distinta de `BOE-A-1992-26318`.
- Cambiar los modelos, las repeticiones, el umbral o los disparadores del job de evals, o la regla con la que una eval informativa deja de decidir.
- Grabar respuestas del BOE distintas de las que necesita la eval nueva, o grabar dentro de un job.
- Cambios en el kernel (`internal/app`, `internal/cli`, `internal/httpx`, `internal/cache`), en el contrato `Applet` o en los esquemas de salida del applet.
- Un ADR nuevo: el mecanismo lo decide el propio hito en `docs/ROADMAP.md` (*Decisión del mecanismo*) y no cambia ninguna decisión de arquitectura.
- Cambios en `docs/SOURCES.md` o en `scripts/verify-sources.sh`: no se toca ninguna fuente nueva ni cómo se consulta.
- Actualizar la entrada de `docs/USO.md` que origina el hito.
- Cambios en el test e2e del applet: su comportamiento visible no cambia.
- Un recuento adicional, en el resultado estructurado o en la forma legible del informe, de en cuántas sesiones cada eval o cada código de aviso se trasladó, aparte de la tasa de sesiones que pasan y del reparto de avisos por sesión que ya exigen FR-040 a FR-042 (Clarifications, sesión 2026-09-16).

## Assumptions

- La tasa de una eval es la que el job ya publica desde ADR 0016: número de sesiones y cuántas pasan. Una eval informativa solo la ejecuta el modelo que decide, con las 3 repeticiones fijadas en el job, de ahí las «tres sesiones» de la aceptación.
- La eval lee y cita el bloque `a42` de `BOE-A-1992-26318` porque es el bloque que H4 ya grabó y el hito lo nombra; como las demás positivas desde T043 de H5, la pregunta nombra la norma y el artículo para medir el protocolo y no la memoria del modelo.
- Los metadatos grabados en H4 de `BOE-A-1992-26318` siguen dando `estatus_derogacion` y `vigencia_agotada` afirmativos, de modo que `articulo` y `metadatos` emiten sobre la caché preparada los avisos `derogada` y `vigencia-agotada`.
- Las etiquetas coinciden con el prefijo de las frases actuales del binario (`⚠ <ETIQUETA>:`), así que exportarlas no cambia ninguna frase.
- El job de evals se dispara al abrir la propuesta de cambio porque el hito toca la skill, sus datos y sus evals (ADR 0016); si hubiera que repetir la ejecución, se relanza a mano o por etiqueta como el job ya permite.
- Dependencias: formato, `Juzgar`, informe, reglas del conjunto, grabaciones y job de evals de H5 (con las enmiendas de ADR 0016), y el applet `boe` de H4 con sus avisos.
