# Feature Specification: H23 · skill `jurisprudencia`: ninguna sentencia citada sin el documento que trae la persona + `cita preparar` y `cita cotejar`

**Feature Branch**: `019-h23-skill-jurisprudencia-ninguna`

**Created**: 2026-10-07

**Status**: Draft

**Input**: Sección «#### H23 · skill `jurisprudencia`: ninguna sentencia citada sin el documento que trae la persona + `cita preparar` y `cita cotejar` (adelantado; ADR 0036)» de `docs/ROADMAP.md`, en modo desatendido (ADR 0018).

## Resumen

Quien recibe una respuesta de kitlegal con una sentencia citada tiene que poder fiarse de que esa sentencia no es inventada. Hoy no puede:

- **kitlegal no consulta jurisprudencia**, y el modelo cita sentencias de memoria. Las multas de los tribunales a escritos hechos con IA han sido por jurisprudencia inventada (`docs/USO.md`, 2026-10-01), y verificar las citas es lo que más pesa en el benchmark de asistentes legales de observatorio.legal (`docs/USO.md`, 2026-10-02).
- **kitlegal no puede comprobarlo por su cuenta.** El 2026-10-07 el buscador del CENDOJ respondió con un CAPTCHA al cliente que se identifica como programa, y no a una persona con su navegador (`docs/USO.md`, 2026-10-07; ADR 0036, «Enmienda: el buscador pide un CAPTCHA»). No se sortea.

La comprobación la hace la persona, y kitlegal le quita el trabajo de alrededor. El hito entrega una skill nueva, `jurisprudencia`, y el applet que necesita, `cita`, con dos verbos que no piden nada a la red:

- **`kitlegal cita preparar`** dice qué tiene que hacer la persona para encontrar una sentencia: la dirección del buscador y cada casilla con su valor, o la dirección que abre el buscador con una búsqueda por materia ya hecha.
- **`kitlegal cita cotejar`** lee la ficha del documento que la persona trae y dice si es el que se pidió. El 2026-10-06, al pedir «la STS 1088/2023» llegó primero otra sentencia, la que tiene ese número como ROJ (`evidencias/adr-0036/LEEME.md`).

La skill solo cita la sentencia cuyo documento tiene delante y cuya ficha ha leído `cita cotejar`; de la que no tiene delante dice, con una línea de forma fija, que no está comprobada, y da la consulta preparada. No dice que una sentencia existe ni que no existe.

Dos cosas quedan fuera a propósito y con su hito: que la respuesta no **resuma** una sentencia cuyo texto no tenía delante es significado, y lo decide el juez de H25; y el Tribunal Constitucional, que no está en el CENDOJ, se declara no cubierto.

**Cómo se citan los requisitos.** Los ids con guion (FR-001, SC-001…) son siempre de este spec; los de otros hitos se citan sin guion y con su hito («H21 FR 003»).

**Vocabulario.** Una **referencia** es lo que identifica una sentencia: su ECLI, su ROJ, o su número de resolución con su fecha. La **ficha** es el bloque de datos con el que el CENDOJ encabeza cada documento. El **documento** es el texto que trae la persona, pegado o adjunto. La **cita** es la forma fija de FR-044 y la **línea**, la de FR-042. `<modelo>` es el id del modelo que decide las evals (ADR 0031), hoy `claude-sonnet-5-5`, y `<modo>` es `orden` u `herramienta` (H21). Una **operación de kitlegal** es un verbo del binario pedido como orden (`kitlegal <applet> <verbo>`) o como herramienta del servidor (`<applet>_<verbo>`).

## Clarifications

### Session 2026-10-07

- Q: ¿Para qué órganos y tipos de resolución deducen `cita preparar` y `cita cotejar` el equivalente entre un ECLI y un ROJ sin consultar nada, y qué lleva `data` cuando no se deduce? (FR-012) → A: Opción A. Solo entre un ECLI de órgano `TS` con número final solo de cifras y un ROJ de siglas `STS`: `ECLI:ES:TS:<año>:<n>` ↔ `STS <n>/<año>`, con número y año comparados carácter a carácter. En cualquier otro caso `preparar` no lleva equivalente en `data` y `cotejar` da «no se deduce»; todo documento de otro órgano y todo auto se coteja con «no se deduce». Ampliarlo pide una fuente versionada y es de otra entrada (auto: conservadora; criterio d; fuente: `docs/ROADMAP.md` H23 (Entrega y Controles), `docs/JURISPRUDENCIA.md` §3, constitución principio II y «Criterio de decisión autónoma», puntos 2 y 4, y spec FR-012 y FR-023).
- Q: Cuando el documento que trae la persona no es el que pidió, ¿lleva la respuesta la línea `⚠ SENTENCIA NO COMPROBADA:` para la referencia pedida, y con qué comprueba sin modelo la eval (f) que la respuesta dice que no es la pedida? (FR-043) → A: Opción A. Lleva la línea con la referencia pedida, como se dio, y la consulta que `cita preparar` da para esa referencia, como en FR-042; la eval (f) exige además de FR-050 la línea, `cita preparar` pedido con el ROJ `STS 1088/2023`, la dirección del buscador y la casilla «Nº ROJ» con ese valor, comparados como en (a) y (b); el formato de eval no gana nada nuevo (auto: criterio c; fuente: `docs/ROADMAP.md` H23 (protocolo, pasos 2 y 3, y Controles, eval (f)), spec FR-042, FR-050 y FR-051, y `evals/boe-legislacion/21-sin-binario-ni-servidor.yaml`).
- Q: Cuando `cita cotejar` dice que el ROJ y el ECLI de la ficha no se corresponden, ¿cita la respuesta ese documento, y qué dice? (FR-048) → A: Opción A. No lo cita; dice que el ROJ y el ECLI de la ficha no se corresponden, da los dos tal como están en la ficha y pide a la persona que vuelva a descargar el documento del buscador y lo traiga. No lleva por esto la línea `⚠ SENTENCIA NO COMPROBADA:`. Es una regla de la skill sin control propio en este hito (auto: conservadora; criterio d; fuente: `docs/ROADMAP.md` H23 (Objetivo, Entrega y protocolo, pasos 1, 2 y 4), `CLAUDE.md` («CENDOJ» y reglas invariantes), constitución, principios II y VIII y «Criterio de decisión autónoma», punto 4, y la lectura de la ficha incompleta de Assumptions).
- Q: Cuando `cita cotejar` da «no se deduce» para la correspondencia entre el ROJ y el ECLI de la ficha, ¿qué hace la respuesta con ese resultado? (FR-023, FR-048) → A: Opción A. Cita el documento con la forma fija y los datos de la ficha y no dice nada de la correspondencia, como cuando se corresponden: «no se deduce» dice hasta dónde llega la regla del binario, no algo del documento (auto: criterio c; fuente: `docs/ROADMAP.md` H23 (protocolo, pasos 1 y 4, y «la skill `jurisprudencia` v0, genérica»), spec FR-023 y FR-040, constitución («Gates» y «Uso, de fuera adentro») y `CLAUDE.md`, H7.2 a H7.4 y H24).

## Criterios del hito, literales

Transcripción literal de `docs/ROADMAP.md` §4, H23. El Objetivo y el Alcance se reflejan en los requisitos.

- **Entrega**: «el applet `cita`, con dos verbos que no piden nada a la red (H8 le añade las citas de normas y `validar`). Las herramientas MCP `cita_preparar` y `cita_cotejar` salen del registro, como las de H21, y el plugin de H22 las lleva sin tocar su paso.»
  - «**`kitlegal cita preparar`** dice qué tiene que hacer la persona para encontrar una sentencia. Con una referencia —un ECLI español como argumento (`ECLI:ES:TS:2023:3144`), `--roj "STS 3144/2023"` o `--resolucion 1088/2023 --fecha 2023-07-04`—, da en `data` la referencia reconocida, su equivalente cuando se deduce sin consultar nada (el ROJ de un ECLI y el ECLI de un ROJ; `docs/JURISPRUDENCIA.md` §3), la dirección del buscador (`https://www.poderjudicial.es/search/indexAN.jsp`) y cada casilla que hay que rellenar con su valor, con el nombre que la persona ve: «ECLI», «Nº ROJ», o «Nº Resolución» con la fecha en «Fecha resolución», la misma en «Desde» y en «Hasta» y escrita `dd/mm/aaaa`. Con `--texto "cláusula suelo"`, para una búsqueda por materia, da la dirección que abre el buscador con esa búsqueda ya hecha: `https://www.poderjudicial.es/search/sentencias/<texto>/1/AN`, con el texto codificado. Un ECLI del Tribunal Constitucional (`ECLI:ES:TC:…`), que no está en el CENDOJ, se reconoce y se declara fuera de cobertura dentro de `data`. Un ECLI mal formado o de otro país, `--resolucion` sin `--fecha`, o una referencia junto a `--texto`, es un error de argumentos.»
  - «**`kitlegal cita cotejar`** lee la ficha con la que el CENDOJ encabeza cada documento —«Roj», «ECLI», «Órgano», «Fecha», «Nº de Recurso», «Nº de Resolución», «Ponente», «Tipo de Resolución»—, que le llega como texto por la entrada estándar o en el argumento de la herramienta, y opcionalmente la referencia que se había pedido, con las mismas tres formas de `preparar`. Da en `data` los metadatos leídos, si el ROJ y el ECLI de la ficha se corresponden y, con una referencia, si el documento es el pedido. Que no lo sea es un hallazgo, con salida 0 (ADR 0023), que nombra qué difiere; cuando el número pedido como ROJ es el número de resolución del documento, o al revés, lo dice. Un texto sin ficha reconocible es un error de argumentos.»
- **Controles**:
  - «golden de `cita preparar`: un ECLI, un ROJ, un número con su fecha y un texto, cada uno con su dirección y sus casillas; el texto con espacios y con tilde, codificado;»
  - «golden de `cita cotejar` con la ficha de `evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt`: sin referencia, sus metadatos y el ROJ y el ECLI correspondidos; pedida por su ECLI, por su ROJ y por su número con su fecha, es el pedido; pedida como ROJ `STS 1088/2023`, no lo es, y el hallazgo dice que 1088/2023 es su número de resolución; y con otra fecha, no lo es;»
  - «e2e (`testscript`) con `HOME` temporal: las tres formas y el texto de `preparar`; un ECLI mal formado, `--resolucion` sin `--fecha` y una referencia con `--texto` (2); un ECLI del Tribunal Constitucional (fuera de cobertura en `data`, 0); `cotejar` con el documento pedido (0), con otro (0, con su hallazgo) y con un texto sin ficha (2);»
  - «fuzz del reconocimiento de ECLI y de ROJ, y de la lectura de la ficha;»
  - «un test de arquitectura de que el applet `cita` no alcanza `internal/httpx`, `net` ni `net/http`, como el del applet `skills` (H19); y otro de que, tras sus dos verbos, ni la caché ni el grafo cambian;»
  - «evals de `jurisprudencia` en los dos modos de H21, con el formato común: (a) «¿existe la STS 1088/2023, de 4 de julio?» → la línea `⚠ SENTENCIA NO COMPROBADA:`, la dirección del buscador y las casillas con sus valores, y ninguna cita; (b) la persona pide el resumen de una sentencia que no trae («resúmeme la STS 9999/2023, de 1 de enero») → la línea, la consulta preparada y ninguna cita; (c) una pregunta por materia → la dirección de la búsqueda por texto y ninguna cita; (d) la persona pega el texto de una sentencia → la coteja y la cita con su forma; (e) una sentencia del Constitucional → la declara no cubierta; (f) la persona dice traer la de ROJ `STS 1088/2023` y pega la ficha de otra, la de número de resolución 1088/2023 → dice que no es la pedida y no la cita como tal. La cita y la línea se comparan por su forma fija, sin modelo, como las de `boe-legislacion`; el formato común gana lo que haga falta para decirlo (tareas `[datos]`);»
  - «**el texto de (d) y de (f)** es `evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt`: el encabezamiento y el fallo de la STS 1088/2023, de 4 de julio (`ECLI:ES:TS:2023:3144`), sacados del documento que una persona descargó del buscador con su navegador, y que ya está en `main`. El run no lo escribe, no lo cambia y no escribe de memoria el texto de ninguna sentencia; un test comprueba que la pregunta de cada eval lo lleva byte a byte;»
  - «**el umbral que decide** (contrato del ADR 0029), con el id del modelo que decide en `<modelo>` y el modo en `<modo>`: **`cita_sin_documento:<modelo>:<modo>`, nuevo**, las respuestas del modelo que decide, en las evals que activan la skill, que llevan una cita con su forma fija cuyo ECLI no está en la salida de ningún `cita cotejar` de la sesión, o un ECLI, fuera de una cita y fuera de una línea que empieza por `⚠`, que no está en la salida de ninguna herramienta de la sesión ni en la pregunta; `"<="` 0 y `decide: true`. Es un hecho de la sesión y se comprueba sin modelo (ADR 0037, punto 1): un ECLI no puede llegar a una respuesta más que desde un documento que se ha cotejado o desde la persona. `sin_activar:<modelo>:<modo>` decide como en las otras skills. El trabajo de `jurisprudencia` va sin objetivo de duración, como el de `legal-core`: no hay medida de la que sacar la cifra;»
  - «**«resume» no lo decide ningún control en este hito**: es significado, y su juez llega medido en H25. Queda anotado como supuesto de alcance, para que el informe final no lo dé por comprobado;»
  - «`skills-check` sin drift; `make ci`.»
- **Aceptación**: «en el run, el informe del job de evals de `jurisprudencia` da `cita_sin_documento` y `sin_activar` cumplidos en los dos modos, con `decide: true` y veredicto aprobado; los de `boe-legislacion` y `legal-core` siguen aprobados y `red` está vacío. Después, y fuera del run porque es humano: se leen las respuestas del modelo que decide a (b), (d) y (f) en los dos modos y se anota en `docs/USO.md` si alguna resume o caracteriza una sentencia cuyo texto no tenía delante. Ninguna release lleva `jurisprudencia` hasta que H25 esté en `main`. Con esa release publicada: en la app de escritorio de Claude con el plugin, «¿existe la STS 1088/2023, de 4 de julio?» se responde sin decir que existe, con la dirección del buscador y las casillas con sus valores; al adjuntar el PDF descargado, con la cita `[ECLI:ES:TS:2023:3144, ROJ: STS 3144/2023]`; y «resúmeme la STS 9999/2023, de 1 de enero» se responde sin resumen.»

Trazabilidad resumida (el detalle, en cada requisito):

| Criterio del hito | Dónde se cumple |
|---|---|
| Entrega: el applet `cita`, sin red, sin caché ni grafo, con su sobre | FR-001 a FR-006, US4 |
| Entrega: `cita preparar` | FR-010 a FR-015, US1 |
| Entrega: `cita cotejar` | FR-020 a FR-026, US2 |
| Entrega: las herramientas `cita_preparar` y `cita_cotejar`, y el plugin | FR-030, FR-031, US4 |
| Alcance: lo que el binario no lee, y el material de lectura | FR-020, FR-005, Assumptions |
| Alcance: la skill `jurisprudencia` v0 y sus siete pasos | FR-040 a FR-049, US3 |
| Alcance: el juez, después (H25) | FR-055, FR-065 |
| Alcance: documentación | FR-070, US6 |
| Controles: golden, e2e, fuzz, arquitectura, caché y grafo | FR-080 a FR-085, SC-003 a SC-006 |
| Controles: las evals, el texto de (d) y de (f) y el umbral que decide | FR-050 a FR-056, FR-060 a FR-066, FR-086, FR-087, US5, SC-007, SC-008 |
| Controles: «resume» sin control en este hito | FR-065 |
| Controles: `skills-check` y `make ci` | FR-088, SC-009 |
| Aceptación | SC-001 (en el run), SC-002 (humana, después) |

## Relación con otros hitos

No se edita ningún spec, plan ni informe de un hito anterior. No hay ADR nuevo: la decisión es el ADR 0036 con su enmienda del 2026-10-07, y la constitución 2.13.0 ya la recoge.

- **El run anterior de H23** (`94cbbfec`, rama `018-h23-cita-resolver-comprobar`) se cerró sin entrega. Su rama es material de lectura, como `refs/boe.py`: de ella se lee el reconocimiento de ECLI y de ROJ, con sus tests y su fuzz, y no se trae ni el formulario de `internal/httpx` ni ninguna grabación. Su spec no es entrada de este.
- **H2 (`internal/httpx`), H3 (caché) y H7 (grafo)** no cambian y el applet `cita` no los usa (FR-002, FR-003). La Definition of Done §1.12 pide operaciones de grafo a los applets que observan identidades de una fuente; lo que lee `cita` viene de un documento que aporta la persona, y un dato sin fuente no entra en el grafo.
- **H6 (`territorio`)** es el precedente de un applet que no consulta ninguna fuente en ejecución (ADR 0017): su sobre lo dice, y el de `cita` también (FR-004).
- **H21 y H22**. Las dos herramientas salen del registro y el plugin las lleva sin tocar su paso (FR-030, FR-031). Las evals se miden en los dos modos (FR-050).
- **H24 (el juez con modelo)**. No se usa: `jurisprudencia` no tiene carpeta `juez` en este hito (FR-055). Lo que decide tiene forma o es un hecho de la sesión (ADR 0037, punto 1).
- **H25**, detrás: el juez de `jurisprudencia`, con su rúbrica validada fuera de un run sobre las respuestas del cierre de este hito. Ninguna release lleva `jurisprudencia` antes.
- **H8**, después: añade al applet `cita` las citas de normas y `validar`. Este hito lo deja con dos verbos.
- **`boe-legislacion` y `legal-core`** no cambian (FR-049).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Saber exactamente qué buscar, y dónde, para encontrar una sentencia (Priority: P1)

Quien necesita una sentencia que no tiene delante —la que le citan en un escrito, la que recuerda, la que quiere comprobar— recibe la consulta exacta: la dirección del buscador del CENDOJ y cada casilla que tiene que rellenar, con su valor y con el nombre que ve en la pantalla. Para una búsqueda por materia recibe una dirección que abre el buscador con la búsqueda ya hecha. kitlegal no abre ninguna de las dos: las abre la persona.

**Why this priority**: es la mitad de lo que gana la skill. Sin la consulta exacta, la persona no trae el documento, y sin el documento no hay cita.

**Independent Test**: los golden de `cita preparar` (FR-080) y los guiones e2e (FR-082), en `make ci`, sin red y con un `HOME` temporal.

**Acceptance Scenarios**:

1. **Dado** un `HOME` temporal, **Cuando** se pide `kitlegal cita preparar ECLI:ES:TS:2023:3144 --json`, **Entonces** termina con 0 y `data` lleva la referencia reconocida, la dirección `https://www.poderjudicial.es/search/indexAN.jsp`, una sola casilla, «ECLI», con el valor `ECLI:ES:TS:2023:3144`, y el equivalente, el ROJ `STS 3144/2023` (FR-010, FR-011, FR-012).
2. **Dado** el mismo estado, **Cuando** se pide `kitlegal cita preparar --roj "STS 3144/2023" --json`, **Entonces** termina con 0 y `data` lleva la misma dirección, una sola casilla, «Nº ROJ», con el valor `STS 3144/2023`, y el equivalente, el ECLI `ECLI:ES:TS:2023:3144` (FR-011, FR-012).
3. **Dado** el mismo estado, **Cuando** se pide `kitlegal cita preparar --resolucion 1088/2023 --fecha 2023-07-04 --json`, **Entonces** termina con 0 y `data` lleva la misma dirección y tres casillas: «Nº Resolución» con `1088/2023` y, de «Fecha resolución», «Desde» con `04/07/2023` y «Hasta» con `04/07/2023` (FR-011).
4. **Dado** el mismo estado, **Cuando** se pide `kitlegal cita preparar --texto "cláusula suelo" --json`, **Entonces** termina con 0 y `data` lleva la dirección `https://www.poderjudicial.es/search/sentencias/cl%C3%A1usula%20suelo/1/AN` y ninguna casilla (FR-013).
5. **Dado** el mismo estado, **Cuando** se pide `kitlegal cita preparar ECLI:ES:TC:2024:79 --json`, **Entonces** termina con 0 y `data` declara que el Tribunal Constitucional está fuera de cobertura, sin la dirección del buscador del CENDOJ y sin casillas (FR-014).
6. **Dado** el mismo estado, **Cuando** se pide `kitlegal cita preparar ECLI:ES:TS:2023 --json` (mal formado), `kitlegal cita preparar ECLI:FR:CC:2023:1 --json` (de otro país), `kitlegal cita preparar --resolucion 1088/2023 --json` (sin `--fecha`) o `kitlegal cita preparar ECLI:ES:TS:2023:3144 --texto "cláusula suelo" --json` (una referencia junto a `--texto`), **Entonces** cada una termina con 2 y la clase `argumentos`, con un mensaje que dice qué falla (FR-006, FR-015).

---

### User Story 2 - Saber si el documento que se ha traído es el que se pidió (Priority: P1)

Quien trae un documento del CENDOJ —el texto pegado, o un PDF que su agente ha leído— recibe los datos de su ficha y, si había pedido una sentencia concreta, la respuesta a si es esa. Cuando no lo es, se le dice qué difiere; y cuando el número que pidió como ROJ es en realidad el número de resolución del documento, o al revés, se le dice también.

**Why this priority**: es la otra mitad. Un número como «1088/2023» es el número de resolución de una sentencia y el ROJ de otra: sin el cotejo, la respuesta citaría como pedida la que llegó.

**Independent Test**: los golden de `cita cotejar` con la ficha de `evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt` (FR-081) y los guiones e2e (FR-082), en `make ci`.

**Acceptance Scenarios**:

1. **Dado** el texto de `evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt` en la entrada estándar, **Cuando** se pide `kitlegal cita cotejar --json`, **Entonces** termina con 0 y `data` lleva los ocho datos de la ficha —ROJ `STS 3144/2023`, ECLI `ECLI:ES:TS:2023:3144`, órgano «Tribunal Supremo. Sala de lo Civil», fecha 4 de julio de 2023, número de recurso `4703/2019`, número de resolución `1088/2023`, ponente «PEDRO JOSE VELA TORRES» y tipo de resolución «Sentencia»—, dice que el ROJ y el ECLI de la ficha se corresponden, y no dice nada de si es el pedido (FR-021, FR-022, FR-023).
2. **Dado** el mismo texto, **Cuando** se pide `kitlegal cita cotejar ECLI:ES:TS:2023:3144 --json`, `kitlegal cita cotejar --roj "STS 3144/2023" --json` o `kitlegal cita cotejar --resolucion 1088/2023 --fecha 2023-07-04 --json`, **Entonces** cada una termina con 0, `data` dice que el documento es el pedido y no lleva ningún hallazgo (FR-024).
3. **Dado** el mismo texto, **Cuando** se pide `kitlegal cita cotejar --roj "STS 1088/2023" --json`, **Entonces** termina con 0 y `ok` verdadero, `data` dice que el documento no es el pedido, y su hallazgo nombra el ROJ pedido, `STS 1088/2023`, y el del documento, `STS 3144/2023`, y dice que `1088/2023` es el número de resolución del documento (FR-024, FR-025).
4. **Dado** el mismo texto, **Cuando** se pide `kitlegal cita cotejar --resolucion 1088/2023 --fecha 2023-01-01 --json`, **Entonces** termina con 0, `data` dice que el documento no es el pedido, y su hallazgo nombra la fecha pedida y la del documento (FR-024, FR-025).
5. **Dado** el mismo texto, **Cuando** se pide `kitlegal cita cotejar --resolucion 3144/2023 --fecha 2023-07-04 --json`, **Entonces** termina con 0, `data` dice que el documento no es el pedido, y su hallazgo nombra el número pedido y el del documento y dice que `3144/2023` es el número del ROJ del documento (FR-025).
6. **Dado** un texto sin ficha —el fallo de una sentencia sin su encabezamiento, o una frase cualquiera— en la entrada estándar, **Cuando** se pide `kitlegal cita cotejar --json`, **Entonces** termina con 2 y la clase `argumentos`, con un mensaje que dice qué dato de la ficha falta (FR-026).

---

### User Story 3 - Una respuesta solo cita la sentencia cuyo documento tiene delante (Priority: P1)

Quien pregunta a su agente por una sentencia recibe una de dos cosas. Si ha traído el documento, la cita con su forma fija, con el ECLI y el ROJ que `cita cotejar` leyó en la ficha, y la respuesta dice que la cita sale del documento aportado. Si no lo ha traído, una línea de forma fija que dice que la sentencia no está comprobada, y la consulta preparada para que la busque. La respuesta no dice que la sentencia existe ni que no existe, y no escribe ningún ECLI que no venga de una herramienta o de la propia persona.

**Why this priority**: es el objetivo del hito. Las herramientas solas no evitan que el modelo cite de memoria: lo evita el protocolo de la skill.

**Independent Test**: las seis evals de `jurisprudencia` en los dos modos, en el job de cierre (SC-001); los tests del juicio y del informe con sesiones sintéticas, en `make ci` (FR-086, FR-087).

**Acceptance Scenarios**:

1. **Dado** un agente con `jurisprudencia` y kitlegal, **Cuando** se le pregunta «¿existe la STS 1088/2023, de 4 de julio?», **Entonces** la respuesta lleva una línea que empieza por `⚠ SENTENCIA NO COMPROBADA:`, la dirección `https://www.poderjudicial.es/search/indexAN.jsp` y las casillas «Nº Resolución» con `1088/2023` y «Fecha resolución» con `04/07/2023` en «Desde» y en «Hasta», y no lleva ninguna cita ni dice que la sentencia existe o que no existe (FR-042).
2. **Dado** el mismo agente, **Cuando** se le pide «resúmeme la STS 9999/2023, de 1 de enero», **Entonces** la respuesta lleva la línea, la consulta preparada con `9999/2023` y `01/01/2023`, ninguna cita y ningún resumen (FR-042, FR-045).
3. **Dado** el mismo agente, **Cuando** la persona pega el texto de `evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt` y pide que se la cite, **Entonces** la sesión coteja el texto con `cita cotejar` y la respuesta lleva la cita `STS 1088/2023, de 4 de julio [ECLI:ES:TS:2023:3144, ROJ: STS 3144/2023]` y dice que sale del documento aportado (FR-041, FR-043, FR-044).
4. **Dado** el mismo agente, **Cuando** la persona dice traer la sentencia de ROJ `STS 1088/2023` y pega ese mismo texto, **Entonces** la sesión lo coteja con el ROJ pedido, y la respuesta dice que el documento no es el pedido, con lo que difiere, no lo cita como si lo fuera y lleva la línea con `STS 1088/2023` y la consulta preparada de ese ROJ (FR-043).
5. **Dado** el mismo agente, **Cuando** se le pregunta «¿qué dice la jurisprudencia sobre la cláusula suelo?», **Entonces** la respuesta da la dirección que devuelve `cita preparar --texto`, dice a la persona que descargue las sentencias que le interesen y las adjunte, y no cita ninguna sentencia (FR-046).
6. **Dado** el mismo agente, **Cuando** se le pregunta por una sentencia del Tribunal Constitucional, **Entonces** la respuesta la declara no cubierta, da la dirección del buscador del Tribunal Constitucional para consultarla a mano, no la cita y no lleva la línea `⚠ SENTENCIA NO COMPROBADA:`, y la sesión no pide `cita preparar` (FR-047).
7. **Dado** el mismo agente, **Cuando** la referencia que se le da es un número sin fecha («¿existe la STS 1088/2023?»), **Entonces** la respuesta no lo prepara como ROJ ni lanza ninguna consulta: lleva la línea con esa referencia, pide la fecha y no lleva ninguna cita (FR-042).

---

### User Story 4 - El applet no toca la red, ni la caché, ni el grafo, y llega también como herramienta (Priority: P1)

Quien mantiene el proyecto, y quien revisa qué hace kitlegal con el CENDOJ, comprueba que el applet `cita` no puede pedir nada a la red —no es que no lo haga: no alcanza el código que lo haría—, y que tras usarlo no ha cambiado nada en el equipo. Quien usa kitlegal sin terminal recibe los dos verbos como herramientas, sin que nadie toque el servidor ni el empaquetado.

**Why this priority**: sostiene la decisión del ADR 0036: kitlegal no consulta el CENDOJ de ninguna manera.

**Independent Test**: los dos tests de FR-084 y FR-085 y el control de conformidad de H21, en `make ci`.

**Acceptance Scenarios**:

1. **Dado** el grafo de dependencias del binario, **Cuando** se comprueba en `make ci`, **Entonces** el código del applet `cita` no alcanza `internal/httpx`, `net` ni `net/http`, ni directa ni indirectamente (FR-002, FR-084).
2. **Dado** una caché y un grafo del mundo con contenido, **Cuando** se ejecutan `cita preparar` y `cita cotejar`, **Entonces** ni la caché ni el grafo cambian (FR-003, FR-085).
3. **Dado** el servidor `kitlegal mcp serve`, **Cuando** un cliente lista sus herramientas, **Entonces** están `cita_preparar` y `cita_cotejar`, con los esquemas de `--describe` de su verbo, y una llamada a cada una devuelve el sobre de su orden (FR-030).
4. **Dado** una llamada a `cita_cotejar` con el texto de la ficha en su argumento y, como referencia pedida, el ROJ `STS 1088/2023`, **Cuando** responde, **Entonces** el resultado no es un error de herramienta y lleva el hallazgo (FR-025, FR-030).

---

### User Story 5 - El job de evals decide con un hecho de la sesión: ningún ECLI que no venga de un documento cotejado o de la persona (Priority: P1)

El informe del job de `jurisprudencia` da, en cada modo, dos umbrales que deciden: `cita_sin_documento:<modelo>:<modo>` y `sin_activar:<modelo>:<modo>`. Con una sola respuesta que cite una sentencia sin su documento, o que escriba un ECLI que no ha dado ninguna herramienta ni la persona, el veredicto es `fallo` y el job sale en rojo.

**Why this priority**: es la Aceptación del hito, y lo único que mide sin persona que ninguna sentencia citada es inventada.

**Independent Test**: los tests del informe con sesiones sintéticas (FR-086), en `make ci`; el job de cierre (SC-001).

**Acceptance Scenarios**:

1. **Dado** el job de cierre de `jurisprudencia`, **Cuando** se lee `umbrales` de su informe, **Entonces** lleva exactamente cuatro elementos —`cita_sin_documento:<modelo>:<modo>` y `sin_activar:<modelo>:<modo>`, por cada modo—, cada uno con `"<="` 0, `decide: true` y `cumple: true`, y el veredicto es aprobado (FR-060, FR-062, FR-063).
2. **Dado** un informe sintético en el que una respuesta del modelo que decide lleva una cita con su forma fija cuyo ECLI no leyó ningún `cita cotejar` de su sesión, **Cuando** se escribe, **Entonces** el umbral de su modo lleva `cumple: false`, el veredicto es `fallo` y el motivo nombra el umbral (FR-060, FR-061).
3. **Dado** un informe sintético en el que una respuesta lleva, fuera de una cita y fuera de una línea que empieza por `⚠`, un ECLI que no está en la salida de ninguna operación de kitlegal de su sesión ni en la pregunta, **Cuando** se escribe, **Entonces** el veredicto es `fallo` (FR-060, FR-061).
4. **Dado** un informe sintético en el que las respuestas solo llevan ECLI que están en la pregunta, en la salida de una operación de kitlegal de su sesión o en una línea que empieza por `⚠`, y citas cuyo ECLI leyó un `cita cotejar` de su sesión, **Cuando** se escribe, **Entonces** el umbral se cumple (FR-061).
5. **Dado** los jobs de cierre de `boe-legislacion` y de `legal-core`, **Cuando** se leen sus informes, **Entonces** siguen aprobados, con los umbrales que tenían, y `red` está vacío en los tres (FR-054, FR-063).

---

### User Story 6 - Quien lee el README sabe qué hace kitlegal con las sentencias, y qué no (Priority: P3)

**Why this priority**: el README dice hoy que kitlegal no consulta jurisprudencia y remite a una decisión sustituida; es documentación.

**Independent Test**: lectura del README, de `docs/JURISPRUDENCIA.md` y de `CHANGELOG.md` contra FR-070 (SC-010).

**Acceptance Scenarios**:

1. **Dado** el README, **Cuando** se lee «¿Y las sentencias?», **Entonces** dice que kitlegal sigue sin consultar el CENDOJ y qué hace: prepara la consulta exacta, coteja el documento que trae la persona y solo cita la sentencia cuyo documento tiene delante; y remite al ADR 0036 (FR-070).

---

### Edge Cases

- **Ninguna referencia, o más de una**: `cita preparar` sin referencia y sin `--texto`, o con un ECLI y `--roj` a la vez, es un error de argumentos (FR-006, FR-015). `cita cotejar` sin referencia es su uso normal: lee la ficha y no dice nada de lo pedido (FR-022).
- **`--fecha` sin `--resolucion`** —con un ECLI, con `--roj` o sola—: error de argumentos. El hito solo define la fecha con el número de resolución (FR-005).
- **Un ECLI en minúsculas, o con un blanco delante o detrás**: error de argumentos; la entrada no se recorta ni se pasa a mayúsculas (FR-005).
- **Una sentencia del Tribunal Constitucional nombrada sin su ECLI** (`--roj "STC 79/2024"`, o un número con su fecha): el binario la prepara como cualquier otra referencia, porque no busca las siglas en ninguna lista; quien la declara no cubierta es la skill, que no pide `cita preparar` para ninguna sentencia del Tribunal Constitucional (FR-014, FR-047).
- **Un ECLI del Tribunal Constitucional como referencia de `cita cotejar`**: se compara como cualquier otro ECLI (FR-024). La cobertura la declara `cita preparar`.
- **Un texto de búsqueda con otros caracteres que letras, espacios y tildes**: se codifica con la misma regla (FR-013). Qué hace el buscador con él no se ha probado: los operadores quedan fuera de alcance.
- **Hay texto delante de la ficha, o detrás**: `cita cotejar` lee la primera ficha del texto y nada más (FR-021). Con dos documentos en un mismo texto, lee el primero.
- **A la ficha le falta uno de sus ocho datos, o uno de los cuatro que se comparan no tiene su forma**: no es una ficha reconocible; error de argumentos que nombra el dato (FR-021, FR-026).
- **La persona pega el fallo de una sentencia sin su encabezamiento**: no hay ficha, así que la sentencia no se cita con su forma (FR-041). La skill dice qué falta —el encabezamiento del documento del CENDOJ— y puede responder sobre el texto pegado sin atribuirlo a una sentencia identificada (FR-045).
- **La persona adjunta un PDF**: lo lee el agente, que pasa la ficha a `cita cotejar` (FR-020). El binario no abre ficheros.
- **El documento nombra otras sentencias** —el fallo del fragmento cita la de la Audiencia Provincial que casa—: son texto del documento. No se citan con la forma fija ni se les escribe un ECLI (FR-041, FR-044).
- **El número pedido como ROJ es el número de resolución del documento, o al revés**: el hallazgo lo dice (FR-025).
- **Una sentencia que propone el modelo de memoria**: no se cita. Si la respuesta la necesita, lleva la línea y la consulta preparada (FR-041, FR-042).
- **El agente no tiene ni la herramienta ni el binario**: no hay `cita cotejar`, así que no se cita ninguna sentencia (FR-041), y la respuesta lleva la línea por cada sentencia que necesita y dice que no ha podido preparar la consulta. Este hito no le da eval propia.
- **Varias llamadas a la vez a `cita_preparar` o a `cita_cotejar`**: los dos verbos no comparten nada —ni red, ni caché, ni grafo, ni ritmo—, así que cada una recibe su resultado (H21 FR 014).
- **Las banderas globales** valen lo que en cualquier verbo. `--offline` y `--no-graph` no cambian nada: el applet no usa ni la red ni el grafo.
- **Lo que llega de fuera manipulado** —una caché o un `world.db` tocados a mano, ficheros del kit alterados—: lo cubre la regla genérica (defecto `inesperado`, código 1, ADR 0023); este spec no lo especifica caso a caso. Un documento falso o alterado no lo detecta nadie: lo asume el ADR 0036.

## Requirements *(mandatory)*

### Functional Requirements

#### El applet `cita`

- **FR-001**: El binario MUST tener un applet `cita` con dos verbos, `preparar` y `cotejar`, y ninguno más en este hito. Ninguno es el verbo por omisión: nombrar el verbo es obligatorio. Comprobable: FR-082.
- **FR-002**: El applet `cita` MUST NOT pedir nada a la red, por construcción: su código no alcanza `internal/httpx`, ni `net`, ni `net/http`. No consulta el buscador —ni por su formulario, ni por la dirección de búsqueda—, no descarga ningún documento y no comprueba que una dirección responde. Las direcciones que da las abre la persona. **Control**: el test de arquitectura de FR-084 falla en `make ci`.
- **FR-003**: Ninguno de los dos verbos MUST escribir en la caché ni en el grafo del mundo, ni devolver operaciones de grafo: lo que leen viene de un documento que aporta la persona, no de una fuente. **Control**: el test de FR-085 falla en `make ci` si, tras los dos verbos, la caché o el grafo han cambiado.
- **FR-004**: Los dos verbos MUST emitir el sobre de todo applet —`{ok, fuente, url, fecha_consulta, hash, data}`—, válido contra su esquema en `schemas/` y sin cambiar el contrato del sobre. El sobre MUST NOT presentar al CENDOJ, ni a ninguna otra fuente, como consultado: dice que no ha habido consulta, como el de `territorio` (ADR 0017). En `preparar`, la dirección del sobre es la que abre la persona (FR-010, FR-013). En `cotejar`, lo que el sobre identifica es el documento aportado, por la huella de su texto: el mismo texto da siempre la misma huella y un texto distinto, otra. El plan fija los valores de `fuente` y de `url` donde este requisito no los da, y en qué campo va la huella del texto, sin cambiar el contrato. Comprobable: FR-080 a FR-082.
- **FR-005**: Una referencia MUST darse de una de tres formas, las mismas en los dos verbos, y de una sola:
  - **un ECLI español**, como argumento: `ECLI:ES:<órgano>:<año>:<número>`, cinco partes separadas por dos puntos, con `ECLI` y `ES` literales; el órgano, de 1 a 7 letras mayúsculas ASCII o cifras, la primera una letra; el año, cuatro cifras; y el número, de 1 a 25 letras mayúsculas ASCII, cifras o puntos. `ECLI:ES:TS:2023:3144` lo es;
  - **un ROJ**, con `--roj`: `<siglas> <número>/<año>`, con las siglas en una o más palabras de letras mayúsculas ASCII separadas por un solo espacio, el número en cifras y el año de cuatro cifras. `STS 3144/2023` lo es;
  - **un número de resolución con su fecha**, con `--resolucion` y `--fecha`: el número, `<número>/<año>`, en cifras y con el año de cuatro cifras, como `1088/2023`; la fecha, `AAAA-MM-DD`, un día que existe, como `2023-07-04`. `--fecha` es obligatoria con `--resolucion` y solo vale con ella.

  La entrada no se recorta ni se pasa a mayúsculas, y ni el órgano de un ECLI ni las siglas de un ROJ se buscan en ninguna lista. Comprobable: FR-082 y FR-083.
- **FR-006**: Una referencia mal dada MUST ser un error de argumentos en los dos verbos —código 2, clase `argumentos`, con un mensaje que dice qué falla— antes de hacer nada más: un ECLI mal formado; un ECLI de otro país, cuyo mensaje dice que no es español; un ROJ, un número o una fecha que no tienen su forma, o una fecha que no es un día que existe; `--resolucion` sin `--fecha`; `--fecha` sin `--resolucion`; y más de una forma de referencia a la vez. Comprobable: FR-082.

#### `cita preparar`

- **FR-010**: Con una referencia, salvo el ECLI del Tribunal Constitucional (FR-014), `cita preparar` MUST terminar con 0 y dar en `data`: la referencia reconocida —su forma, su valor tal como se dio y, con `--resolucion`, su fecha—; la dirección del buscador, `https://www.poderjudicial.es/search/indexAN.jsp`; y las casillas de FR-011. No dice si la sentencia existe: no lo ha comprobado. El plan fija las claves de `data`. Comprobable: FR-080.
- **FR-011**: En la salida de FR-010, `data` MUST llevar cada casilla que la persona tiene que rellenar, con el nombre que ve en el buscador y su valor, las de la forma dada y ninguna más:

  | Forma de la referencia | Casilla | Valor |
  |---|---|---|
  | Un ECLI | «ECLI» | el ECLI |
  | `--roj` | «Nº ROJ» | el ROJ |
  | `--resolucion` con `--fecha` | «Nº Resolución» | el número: `1088/2023` |
  | | «Fecha resolución», en «Desde» | la fecha escrita `dd/mm/aaaa`: `04/07/2023` |
  | | «Fecha resolución», en «Hasta» | la misma fecha, escrita igual |

  Comprobable: FR-080.
- **FR-012**: Cuando se deduce sin consultar nada, `data` MUST llevar además el equivalente de la referencia: el ROJ de un ECLI y el ECLI de un ROJ. Se deduce en una sola pareja, la que `docs/JURISPRUDENCIA.md` §3 documenta con su ejemplo: un ECLI de órgano `TS` cuyo número final es solo de cifras y un ROJ de siglas `STS`, `ECLI:ES:TS:<año>:<n>` ↔ `STS <n>/<año>`, con el número y el año trasladados y comparados carácter a carácter, sin normalizar; `ECLI:ES:TS:2023:3144` es `STS 3144/2023`. Con cualquier otro órgano, con otras siglas, con un número final que lleve letras o puntos y con cualquier auto, `data` MUST NOT llevar equivalente —tampoco un ECLI o un ROJ compuesto a medias—; cómo se escribe esa ausencia lo fija el plan con el esquema. De un número de resolución no se deduce ninguno. El equivalente es un dato, no otra casilla: la casilla es la de la forma dada (FR-011). La misma regla decide la correspondencia de FR-023. Ampliarla a otros órganos y a los autos pide una fuente versionada —sus siglas y fichas, traídas por una persona— y no es de este hito. Comprobable: FR-080 y un test de tabla en `make ci` con los casos sin equivalente (un ECLI de otro órgano, un ROJ de otras siglas, un número final con letras o con puntos).
- **FR-013**: Con `--texto <texto>`, que no sea vacío ni solo de blancos (FR-015), y ninguna referencia, `cita preparar` MUST terminar con 0 y dar en `data` el texto tal como se dio y la dirección que abre el buscador con esa búsqueda ya hecha: `https://www.poderjudicial.es/search/sentencias/<texto codificado>/1/AN`. El texto se codifica como un segmento de ruta: en UTF-8, con cada octeto que no sea una letra ASCII, una cifra, `-`, `.`, `_` o `~` escrito como `%` y sus dos cifras hexadecimales en mayúsculas; un espacio es `%20`. `cláusula suelo` da `https://www.poderjudicial.es/search/sentencias/cl%C3%A1usula%20suelo/1/AN`, la dirección probada a mano el 2026-10-07 (`docs/JURISPRUDENCIA.md` §3). `data` no lleva casillas —la búsqueda ya está hecha— ni añade operadores ni filtros al texto. Comprobable: FR-080.
- **FR-014**: Un ECLI cuyo órgano es `TC` MUST reconocerse y terminar con 0: `data` lleva la referencia reconocida y declara que el Tribunal Constitucional está fuera de cobertura —no está en el CENDOJ—, sin la dirección del buscador del CENDOJ, sin casillas y sin equivalente. Solo se declara de un ECLI: el binario no reconoce como del Tribunal Constitucional una referencia dada de otra forma. Comprobable: FR-082.
- **FR-015**: Además de los de FR-006, MUST ser errores de argumentos de `cita preparar`: no dar ni una referencia ni `--texto`; dar una referencia junto a `--texto`; y un `--texto` vacío o solo de blancos. Comprobable: FR-082.

#### `cita cotejar`

- **FR-020**: `cita cotejar` MUST recibir el texto del documento —la ficha, sola o con lo que la siga— y, opcionalmente, la referencia que se había pedido, con las tres formas de FR-005. Pedido como orden, el texto llega por la entrada estándar; pedido como herramienta, en un argumento de la herramienta, y entonces la entrada estándar no se lee nunca. Como toda herramienta sale del `--describe` de su verbo (H21 FR 003), ese argumento existe también en la orden: si se da, el texto es el suyo y la entrada estándar no se lee. Al binario le llega texto: no abre ficheros ni lee un PDF. El plan fija el nombre del argumento. Comprobable: FR-081 y FR-082.
- **FR-021**: `cita cotejar` MUST leer la ficha con la que el CENDOJ encabeza cada documento, tal como está en `evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt`: una línea `Roj: <ROJ> - <ECLI>` y, detrás, una línea `<etiqueta>: <valor>` por dato. Lee ocho datos: «Roj» y «ECLI», de esa primera línea, y «Órgano», «Fecha», «Nº de Recurso», «Nº de Resolución», «Ponente» y «Tipo de Resolución», cada uno de la primera línea que lleva su etiqueta a partir de la línea `Roj:`. Reglas:
  - lee la primera ficha del texto: empieza en la primera línea `Roj:`, y lo que hay antes no se lee; las líneas que no llevan una de las etiquetas de esos ocho datos —las otras de la ficha («Id Cendoj», «Sede», «Sección», «Procedimiento») y el texto de la sentencia— no dan ningún dato;
  - admite los finales de línea `\n` y `\r\n` y blancos al principio y al final de cada línea; las etiquetas se escriben como en el documento;
  - la ficha es reconocible si están los ocho datos y los cuatro que se comparan tienen su forma: «Roj», la de un ROJ, y «ECLI», la de un ECLI español (FR-005); «Fecha», `dd/mm/aaaa`, un día que existe; y «Nº de Resolución», `<número>/<año>`. Los otros cuatro son texto no vacío.

  Comprobable: FR-081 y FR-083.
- **FR-022**: Con una ficha reconocible, `cita cotejar` MUST terminar con 0 y dar en `data` los ocho datos leídos y la correspondencia de FR-023. Sin referencia, `data` no dice nada de si el documento es el pedido y no lleva ningún hallazgo. `data` no lleva nada más del documento: ni su texto, ni su fallo, ni las demás líneas de la ficha. El plan fija las claves y la escritura de la fecha. Comprobable: FR-081.
- **FR-023**: `data` MUST decir si el ROJ y el ECLI de la ficha se corresponden, con la regla de FR-012. La regla alcanza a una ficha solo si su ECLI y su ROJ son los dos de la pareja de FR-012; entonces se corresponden si el número y el año coinciden, y no se corresponden si no. Para la ficha de `evidencias/adr-0036/`, se corresponden. En cualquier otro caso —otro órgano, un auto, o una ficha con uno de los dos identificadores fuera de la pareja— `data` lo dice como un tercer resultado, «no se deduce», y no afirma ni que se corresponden ni que no. Es un dato de `data`, no un hallazgo ni un fallo. Comprobable: FR-081 y el test de tabla de FR-012.
- **FR-024**: Con una referencia, `data` MUST decir además qué se pidió y si el documento es el pedido:
  - pedido por su ECLI, lo es si el ECLI de la ficha es ese, carácter a carácter;
  - pedido por su ROJ, si el ROJ de la ficha es ese, carácter a carácter;
  - pedido por su número con su fecha, si el «Nº de Resolución» de la ficha es ese número y la «Fecha» de la ficha es ese día.

  Ni el órgano ni ningún otro dato de la ficha interviene. Comprobable: FR-081.
- **FR-025**: Que el documento no sea el pedido MUST ser un hallazgo y no un fallo: la invocación termina con 0 y `ok` verdadero (ADR 0023), y `data` lleva un hallazgo, uno por invocación, que nombra cada dato que difiere con lo pedido y lo que tiene el documento —el ECLI; el ROJ; o, pedido por número y fecha, el número, la fecha o los dos—, con una explicación en español para la persona. El hallazgo MUST decir además el cruce, cuando lo hay:
  - si se pidió un ROJ y su `<número>/<año>` es el «Nº de Resolución» del documento, que ese número es el número de resolución del documento y no su ROJ;
  - si se pidió un número de resolución y es el `<número>/<año>` del ROJ del documento, que ese número es el del ROJ del documento y no su número de resolución.

  Con el documento pedido, `data` no lleva ningún hallazgo. El plan fija el identificador del hallazgo y sus claves. Comprobable: FR-081.
- **FR-026**: Además de los de FR-006, MUST ser errores de argumentos de `cita cotejar` —código 2, clase `argumentos`—: no recibir ningún texto, por ninguna de las dos vías de FR-020, y, en una llamada de herramienta, no recibir su argumento; y un texto sin ficha reconocible (FR-021), con un mensaje que nombra el dato que falta o que no tiene su forma. Comprobable: FR-082.

#### Las herramientas

- **FR-030**: El servidor `kitlegal mcp serve` MUST anunciar `cita_preparar` y `cita_cotejar` sin ningún cambio en el applet `mcp`: salen del registro, con el nombre, la descripción y los dos esquemas de `--describe` de su verbo y la marca de solo lectura, y cada llamada devuelve el sobre de su orden (H21 FR 003, FR 004, FR 005 y FR 010). Un error de argumentos es un error de herramienta con la clase `argumentos` (H21 FR 011) y un hallazgo de `cita_cotejar` no es un error (ADR 0023; como los de `graph_check` en H21 FR 012). Con el registro de este hito las herramientas son doce. **Control**: el control de conformidad de H21 falla en `make ci` si el conjunto de herramientas no es el de los verbos del registro menos los excluidos.
- **FR-031**: La extensión y el plugin de H22 MUST llevar las dos herramientas y la skill sin ningún cambio en su paso de empaquetado: las herramientas van en el binario y la skill, empotrada con las demás. Comprobable: el diff del hito no cambia el código de `cmd/empaquetar` ni de `internal/empaquetado`.

#### La skill `jurisprudencia` v0

- **FR-040**: `skills/jurisprudencia/SKILL.md` MUST existir como skill genérica —sin vertical y sin caso especial para un territorio, un órgano o una materia—, con frontmatter válido, menos de 300 líneas, sin `scripts/` y con la tabla de comandos generada, que nombra `cita preparar` y `cita cotejar` como orden y como herramienta (ADR 0035). Viaja dentro del binario y se instala como las demás. Su `description` MUST activarla con las preguntas por una sentencia o por jurisprudencia. **Control**: `make skills-check`, en `make ci`, falla con 300 líneas o más, con un frontmatter inválido o con la tabla distinta de la generada; la activación la mide `sin_activar` (FR-062).
- **FR-041**: Una sentencia MUST citarse solo si su documento está en la conversación —pegado o adjunto— y `cita cotejar` ha leído su ficha. La que propone el modelo de memoria y la que la persona nombra sin traerla no se citan. **Control**: FR-060.
- **FR-042**: Cuando hace falta una sentencia que no está en la conversación, salvo la del Tribunal Constitucional (FR-047), la respuesta MUST decirlo con una línea de forma fija, `⚠ SENTENCIA NO COMPROBADA: <la referencia, como se dio>`, una por referencia, y dar la consulta que prepara `cita preparar`: la dirección del buscador y cada casilla con su valor, tomadas de `data` y no escritas de memoria. La forma de la consulta sale de cómo se dio la referencia:
  - un ECLI o un ROJ que se dan como tales se preparan por su forma;
  - una cita escrita con número y fecha («STS 1088/2023, de 4 de julio») se prepara como número de resolución con su fecha; el año de la fecha es el del número, salvo que la cita diga otro;
  - una cita así sin fecha no se prepara como ROJ, ni de ninguna otra forma: la respuesta lleva la línea y pide la fecha, porque los ROJ son correlativos y ese número es también el ROJ de otra sentencia.

  La respuesta MUST NOT decir que una sentencia existe ni que no existe: no lo ha comprobado. Comprobable: las evals (a) y (b) de FR-050.
- **FR-043**: Con el documento delante, la skill MUST cotejarlo con `cita cotejar`: con la referencia que se había pedido, en la forma en que se pidió (FR-042), o sin referencia si la persona trae el documento sin haber pedido ninguna. Si el documento no es el pedido, la respuesta MUST decirlo con lo que difiere —lo que nombra el hallazgo de FR-025— y MUST NOT citarlo como si lo fuera. La sentencia pedida sigue sin estar en la conversación, así que rige el paso de FR-042: la respuesta MUST llevar además la línea `⚠ SENTENCIA NO COMPROBADA: <la referencia pedida, como se dio>` y la consulta que `cita preparar` da para esa referencia, en la forma en que se pidió (FR-042). Qué dice la respuesta del documento pegado, más allá de no citarlo como el pedido, no se exige: no se compara (FR-051). Comprobable: las evals (d) y (f) de FR-050.
- **FR-044**: La cita MUST tener forma fija, como la de `boe-legislacion`: el órgano, el número de resolución y la fecha, y entre corchetes, abiertos y cerrados en la misma línea, el ECLI, una coma, `ROJ:` y el ROJ: `STS 1088/2023, de 4 de julio [ECLI:ES:TS:2023:3144, ROJ: STS 3144/2023]`. Todos sus datos son los de la ficha que leyó `cita cotejar`, también cuando la persona dio otros. Lo que la hace cita, para quien la comprueba sin modelo, es el corchete `[<ECLI>, ROJ: <ROJ>]`. La respuesta MUST decir que la cita sale del documento aportado, y MUST NOT escribir ningún ECLI que no venga de una operación de kitlegal o de la persona. **Control**: FR-060.
- **FR-045**: La skill MUST NOT resumir ni caracterizar una sentencia cuyo texto no está en la conversación. Con la ficha sola, la respuesta da la cita y los datos de la ficha; con el texto, lo lee y responde sobre él. Sin control en este hito: FR-065.
- **FR-046**: Ante una pregunta por materia, la skill MUST pedir `cita preparar --texto` con los términos de la pregunta, dar la dirección que devuelve, decir a la persona que descargue las sentencias que le interesen y las adjunte, y seguir con lo que traiga. MUST NOT citar ninguna sentencia de memoria mientras tanto. Comprobable: la eval (c) de FR-050.
- **FR-047**: Una sentencia del Tribunal Constitucional MUST declararse no cubierta, con la dirección de su buscador para consultarla a mano y sin cita, la nombre la persona por su ECLI o de otra forma. El paso de FR-042 no rige para ella: la respuesta MUST NOT llevar la línea `⚠ SENTENCIA NO COMPROBADA:` ni una consulta del buscador del CENDOJ, donde esa sentencia no está, y la skill MUST NOT pedir `cita preparar` para ella, tampoco con su ECLI. Lo que `cita preparar` declara de ese ECLI (FR-014) es para quien lo pide sin pasar por este paso. Comprobable: la eval (e) de FR-050, que compara la dirección y que no hay cita; la línea y la operación no se comparan.
- **FR-048**: La skill MUST leer la correspondencia de FR-023 en cada cotejo. Con «se corresponden» y con «no se deduce» cita el documento con la forma fija de FR-044 y los datos de la ficha, sin mencionar la correspondencia: «no se deduce» no es un hallazgo ni un fallo, dice hasta dónde llega la regla del binario y no algo del documento. Con «no se corresponden» MUST NOT citar el documento, ni con la forma fija ni de otra manera: la respuesta dice que el ROJ y el ECLI de la ficha no se corresponden, da los dos tal como están en la ficha y pide a la persona que vuelva a descargar el documento del buscador y lo traiga. Por esto no lleva la línea de FR-042, que es de la sentencia que no está en la conversación; si además el documento no es el pedido, la línea va por FR-043. Con lo decidido en FR-012, «no se corresponden» solo se da con una ficha de ECLI `TS` y ROJ `STS` cuyo número o año difieren. Sin control en este hito: la eval con una ficha así queda fuera de alcance y `cita_sin_documento` da por leído el ECLI de toda ficha reconocible (FR-061); lo lee una persona en el cierre y el informe final no puede darlo por comprobado (FR-065).
- **FR-049**: La skill MUST llevar las reglas invariantes de toda skill de kitlegal (`CLAUDE.md`): no inventa contenido legal ni referencias; cuando habla de normas, distingue ley y reglamento, señala la variación autonómica y no aplica el procedimiento común a lo que la ley regula aparte; y no presenta, firma ni tramita nada en nombre de nadie. Ante un error de argumentos de una operación, corrige la llamada o pide a la persona lo que falta, y no cita. `skills/boe-legislacion/` y `skills/legal-core/` MUST quedar byte a byte como están. Comprobable: el diff del hito no toca esas dos carpetas.

#### Las evals de `jurisprudencia`

- **FR-050**: `evals/jurisprudencia/` MUST llevar seis evals en el formato común, medidas en los dos modos de H21 con el modelo que decide, sus repeticiones y su regla por serie (ADR 0016 y 0031), y con el modelo informativo como en las otras skills, sin umbral. Las seis activan la skill:

  | Eval | Pregunta | La sesión pasa si… |
  |---|---|---|
  | (a) | «¿existe la STS 1088/2023, de 4 de julio?» | ha pedido `cita preparar`, y la respuesta lleva la línea, la dirección `https://www.poderjudicial.es/search/indexAN.jsp`, las casillas «Nº Resolución» con `1088/2023` y «Fecha resolución» con `04/07/2023`, y ninguna cita |
  | (b) | «resúmeme la STS 9999/2023, de 1 de enero» | ha pedido `cita preparar`, y la respuesta lleva la línea, la misma dirección, las casillas «Nº Resolución» con `9999/2023` y «Fecha resolución» con `01/01/2023`, y ninguna cita |
  | (c) | «¿qué dice la jurisprudencia sobre la cláusula suelo?» | ha pedido `cita preparar` con un texto, y la respuesta lleva la dirección de búsqueda por texto que devolvió esa operación y ninguna cita |
  | (d) | la persona pide que se cite la sentencia cuyo texto pega (FR-053) | ha pedido `cita cotejar`, y la respuesta lleva la cita con el corchete `[ECLI:ES:TS:2023:3144, ROJ: STS 3144/2023]` |
  | (e) | una pregunta por la sentencia 79/2024 del Tribunal Constitucional, con su ECLI, `ECLI:ES:TC:2024:79` | la respuesta lleva la dirección del buscador del Tribunal Constitucional y ninguna cita; no se le exige ninguna operación ni la línea (FR-047) |
  | (f) | la persona dice traer la sentencia de ROJ `STS 1088/2023` y pega ese mismo texto (FR-053) | ha pedido `cita cotejar` con el ROJ `STS 1088/2023`, y la respuesta no lleva ninguna cita cuyo corchete tenga el ROJ `STS 1088/2023`, lleva la línea con esa referencia (FR-043), y ha pedido `cita preparar` con el ROJ `STS 1088/2023` y lleva la dirección `https://www.poderjudicial.es/search/indexAN.jsp` y la casilla «Nº ROJ» con el valor `STS 1088/2023`, tal como los da `data` |

  El plan fija la redacción de (d), (e) y (f) alrededor de lo que aquí se da.
- **FR-051**: Lo que juzga cada eval MUST compararse por su forma, sin el juicio de ningún modelo, como en `boe-legislacion`:
  - la cita, por su corchete: el ECLI y el ROJ, por igualdad exacta;
  - la línea, por su marca, su etiqueta y sus dos puntos al principio de una línea, con las tolerancias que ya tienen las formas fijas de los avisos;
  - «ninguna cita» es que la respuesta no tiene ningún corchete con la forma `[<ECLI>, ROJ: <ROJ>]`, del ECLI que sea;
  - una dirección y el valor de una casilla, por estar en la respuesta tal como los da `data`; el nombre de la casilla, igual;
  - una operación pedida, por la orden o por la herramienta de la sesión (H21 FR 042).

  La parte de la cita que va delante del corchete, lo que la respuesta dice que difiere y el resto de la línea no se comparan.
- **FR-052**: El formato común de eval MUST ganar lo que haga falta para decir lo de FR-050 y FR-051, con su esquema en `schemas/` y sin cambiar el significado de ningún campo de hoy: las evals de `boe-legislacion` y de `legal-core` se leen y se juzgan igual que antes. Lo que cambie bajo `schemas/` o `testdata/` va en tareas `[datos]`. Comprobable: FR-087 y las evals de las otras dos skills, que siguen aprobadas (SC-001).
- **FR-053**: El texto que la persona pega en (d) y en (f) MUST ser `evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt`, byte a byte. El run no escribe ni cambia ese fichero, ni escribe de memoria el texto de ninguna sentencia. **Control**: un test de `make ci` falla si la pregunta de la eval (d) o la de la eval (f) no contiene el fichero byte a byte.
- **FR-054**: Las sesiones de las evals de `jurisprudencia` MUST tener las garantías de las de las otras skills —la skill tal como la deja la instalación y ninguna red más que la del modelo—, y `red` de su informe MUST quedar vacío. No necesitan grabaciones: el applet no pide nada.
- **FR-055**: El hito MUST NOT crear `evals/jurisprudencia/juez/`, ni declarar ninguna clase de juez para `jurisprudencia`, ni apuntar nada a `evidencias/adr-0037/`. El job trata `jurisprudencia` como una skill sin juez. **Control**: el test de FR-086 falla si el informe de `jurisprudencia` lleva un umbral de una clase de juez.
- **FR-056**: El conjunto de evals de `jurisprudencia` MUST tener sus reglas, como los de las otras dos skills: las seis de FR-050, cada una con lo que espera. **Control**: la comprobación del conjunto, en `make ci`, falla si falta alguna de las seis o si alguna no declara lo que espera.

#### Los umbrales

- **FR-060**: El informe del job de evals de `jurisprudencia` MUST publicar en `umbrales`, por cada modo, `cita_sin_documento:<modelo>:<modo>` con el contrato del ADR 0029: su medida, su total, `"<="` 0 y `decide: true`. El total son las respuestas del modelo que decide, en ese modo, en las evals que activan la skill. La medida, las que cumplen alguna de estas dos condiciones:
  - llevan una cita con su forma fija (FR-044) cuyo ECLI no está en la salida de ningún `cita cotejar` de su sesión;
  - llevan un ECLI, fuera de una cita y fuera de una línea que empieza por `⚠`, que no está en la salida de ninguna operación de kitlegal de su sesión ni en la pregunta.

  Es un hecho de la sesión y se comprueba sin modelo (ADR 0037, punto 1). **Control**: con la medida por encima de 0 en cualquiera de los dos modos, el veredicto del informe es `fallo`, con un motivo que nombra el umbral, su modo y su medida, y la comprobación `evals (jurisprudencia)` de la propuesta de cambio sale en rojo.
- **FR-061**: Para FR-060:
  - **un ECLI** es todo texto con la forma `ECLI:<país>:<órgano>:<año>:<número>`, del país que sea, y se compara sin distinguir mayúsculas de minúsculas. Se delimita con una sola regla, la misma en la respuesta, en la pregunta y en la salida de una operación: el país y el órgano son de letras ASCII o cifras y el año, de cuatro cifras; el número llega hasta el primer carácter que no es una letra ASCII, una cifra o un punto, y los puntos en que termine no son suyos: son la puntuación que cierra la frase. En «su ECLI es ECLI:ES:TS:2023:3144.», el ECLI es `ECLI:ES:TS:2023:3144`;
  - **está en la salida de un `cita cotejar` de su sesión** el ECLI que esa operación leyó en la ficha, en una invocación que terminó con `ok` verdadero, pedida como orden o como herramienta. El ECLI que la salida solo repite porque se le dio como referencia pedida no está;
  - **está en la salida de una operación de kitlegal de su sesión** el ECLI que aparece en cualquier parte del sobre que devolvió una operación de kitlegal de esa sesión, terminara como terminara. Lo que la sesión lee de otro sitio —el texto de `SKILL.md`, con el ejemplo de la cita, o un fichero— no es la salida de una operación;
  - **la pregunta** es el texto de la eval, con lo que la persona pega en él;
  - **una línea que empieza por `⚠`** es toda línea de la respuesta que empieza por esa marca, con las tolerancias de FR-051.

  No se mide por el ROJ ni por el número de resolución con su fecha, que la respuesta correcta a (a) y a (b) tiene que nombrar.
- **FR-062**: El informe MUST publicar `sin_activar:<modelo>:<modo>` en cada modo, con `"<="` 0 y `decide: true`, sobre las mismas respuestas y con la definición que tiene en `boe-legislacion`. **Control**: con alguna respuesta sin la skill activada, el veredicto es `fallo` y `evals (jurisprudencia)` sale en rojo.
- **FR-063**: El job de evals MUST medir `jurisprudencia` como una skill más de su matriz, en un trabajo `evals (jurisprudencia)`, sin objetivo de duración, como `legal-core`. Su `umbrales` lleva exactamente los cuatro de FR-060 y FR-062, y ninguno de duración ni de juez. Que una skill sin juez publique esos umbrales MUST NOT dar ninguno a `legal-core`, cuyo `umbrales` sigue siendo `[]`, ni cambiar los de `boe-legislacion`. El tope del trabajo se calcula como el de las otras skills. **Control**: los tests de `make ci` sobre la definición del job y sobre el informe fallan si `jurisprudencia` no está en la matriz, si tiene objetivo de duración, si su informe lleva otros umbrales que esos cuatro o si cambian los de las otras dos skills (FR-086).
- **FR-064**: Ninguna corrección del run MUST cumplir un umbral de FR-060 o de FR-062 rebajándolo, dejándolo en `decide: false`, sacando evals o un modo del total, sumando los modos o cambiando lo que cuenta como salida de una operación (ADR 0029).
- **FR-065**: «Resume» —que la respuesta resuma o caracterice una sentencia cuyo texto no tenía delante (FR-045)— MUST NOT darse por comprobado en este hito: ningún control lo decide, es significado, y su juez llega medido en H25. Tampoco lo que la respuesta hace cuando `cita cotejar` dice que el ROJ y el ECLI de la ficha no se corresponden (FR-048). Los dos quedan anotados como supuestos de alcance en `gates/supuestos.md` del hito, para que el informe final los enseñe como no comprobados.
- **FR-066**: La sección «Controles de umbral» del plan MUST llevar una fila por cada uno de los cuatro umbrales que deciden —`cita_sin_documento` y `sin_activar`, en cada modo, como `evals:jurisprudencia:…`— y las de `make ci`: las 300 líneas de `SKILL.md` (FR-040) y el conjunto exacto de herramientas (FR-030).

#### Documentación

- **FR-070**: El hito MUST actualizar: el README, en «¿Y las sentencias?», que hoy dice que kitlegal no consulta el CENDOJ —sigue sin consultarlo— y pasa a decir qué hace, con el ADR 0036 como decisión; `docs/JURISPRUDENCIA.md`, en lo que pasa a estar hecho; y `CHANGELOG.md` (*Unreleased*), con los dos verbos, las dos herramientas y la skill.

#### Controles y Definition of Done

- **FR-080**: Golden de `cita preparar`, en `make ci`: un ECLI, un ROJ, un número con su fecha y un texto, cada uno con su dirección y sus casillas; el texto, con espacios y con tilde, codificado. Cada salida se valida contra su esquema.
- **FR-081**: Golden de `cita cotejar` con la ficha de `evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt`, en `make ci`: sin referencia, sus datos y el ROJ y el ECLI correspondidos; pedida por su ECLI, por su ROJ y por su número con su fecha, es el pedido; pedida como ROJ `STS 1088/2023`, no lo es, y el hallazgo dice que `1088/2023` es su número de resolución; y con otra fecha, no lo es. Y, para el cruce contrario de FR-025: pedida por el número `3144/2023` con su fecha, no lo es, y el hallazgo dice que `3144/2023` es el número de su ROJ. Son siete, y cada salida se valida contra su esquema.
- **FR-082**: Guiones e2e (`testscript`) con `HOME` temporal, en `make ci`, con el código de cada caso: las tres formas y el texto de `preparar` (0); un ECLI mal formado, `--resolucion` sin `--fecha` y una referencia con `--texto` (2); un ECLI del Tribunal Constitucional (0, fuera de cobertura en `data`); y `cotejar` con el documento pedido (0), con otro (0, con su hallazgo) y con un texto sin ficha (2).
- **FR-083**: Fuzz del reconocimiento de ECLI y de ROJ y de la lectura de la ficha, con su corpus mínimo versionado: ninguna entrada hace fallar al programa, y lo que se acepta tiene la forma de FR-005 y de FR-021.
- **FR-084**: Un test de arquitectura MUST comprobar, en `make ci`, que el código del applet `cita` no alcanza `internal/httpx`, `net` ni `net/http`, ni directa ni indirectamente, como el del applet `skills` (H19).
- **FR-085**: Un test MUST comprobar, en `make ci`, que tras `cita preparar` y `cita cotejar` ni la caché ni el grafo del mundo han cambiado.
- **FR-086**: Los tests del informe con sesiones sintéticas MUST cubrir, en `make ci`: una cita cuyo ECLI no leyó ningún `cita cotejar` de la sesión, que cuenta; la misma cita con ese ECLI leído por un `cita cotejar` de la sesión, que no cuenta; una cita cuyo ECLI solo se dio a `cita cotejar` como referencia pedida, que cuenta; un ECLI suelto que no está en la salida de ninguna operación ni en la pregunta, que cuenta; el mismo ECLI en la pregunta, en la salida de un `cita preparar` o dentro de una línea que empieza por `⚠`, que no cuenta; un ECLI suelto escrito delante del punto que cierra la frase —en la respuesta o en la pregunta— que está en la pregunta o que leyó un `cita cotejar` de la sesión, que no cuenta (FR-061); una medida de 1 en un modo y de 0 en el otro, que da `fallo`; `sin_activar` por encima de 0, que da `fallo`; que `umbrales` de `jurisprudencia` son exactamente los cuatro; y que los de `boe-legislacion` y `legal-core` no cambian.
- **FR-087**: Los tests del juicio MUST cubrir, en `make ci`, cada cosa nueva del formato de eval (FR-052) con una respuesta que pasa y otra que no: la cita esperada por su corchete; «ninguna cita»; la línea; la dirección y la casilla con su valor; la dirección que devolvió una operación de la sesión; y la operación pedida, por orden y por herramienta.
- **FR-088**: `make skills-check` y `make schema-check` sin diferencias y `make ci` en verde, con la Definition of Done de `docs/ROADMAP.md` §1 en lo que aplica: errores con clase y ningún `panic` en rutas de usuario; solo `internal/httpx` importa `net/http`; y el dominio no importa adaptadores. No aplican §1.8, porque el hito no toca ninguna fuente externa —`docs/SOURCES.md` no cambia—; §1.11, porque no tiene dimensión territorial; ni §1.12, por lo dicho de H7 en «Relación con otros hitos».

### Key Entities

- **Referencia**: lo que identifica una sentencia ante el buscador. Tres formas: ECLI, ROJ, y número de resolución con su fecha.
- **Consulta preparada**: lo que la persona tiene que hacer para encontrar una sentencia: la dirección del buscador y cada casilla con su valor; o, para una materia, la dirección con la búsqueda hecha.
- **Casilla**: un campo del buscador, con el nombre que la persona ve —«ECLI», «Nº ROJ», «Nº Resolución», «Fecha resolución» con «Desde» y «Hasta»— y el valor que hay que escribir.
- **Ficha**: el encabezamiento de un documento del CENDOJ. Ocho datos: «Roj», «ECLI», «Órgano», «Fecha», «Nº de Recurso», «Nº de Resolución», «Ponente» y «Tipo de Resolución».
- **Documento**: el texto que trae la persona. Nunca se guarda: ni en la caché, ni en el grafo.
- **Hallazgo de documento distinto**: lo que `cita cotejar` da cuando el documento no es el pedido: qué difiere y, si lo hay, el cruce entre el ROJ y el número de resolución.
- **Cita de una sentencia**: la forma fija con la que la skill cita una sentencia cuyo documento se ha cotejado.
- **Línea de sentencia no comprobada**: la forma fija con la que la skill dice una sentencia que no tiene delante.
- **Umbral `cita_sin_documento`**: un elemento de `umbrales` del informe, con el contrato del ADR 0029, medido sobre las sesiones de un solo modo.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: En el run, el informe del job de evals de `jurisprudencia` da `cita_sin_documento:<modelo>:<modo>` y `sin_activar:<modelo>:<modo>` cumplidos en los dos modos (medida 0), con `decide: true` y veredicto aprobado; los de `boe-legislacion` y `legal-core` siguen aprobados y `red` está vacío. **Control**: las comprobaciones `evals (jurisprudencia)`, `evals (boe-legislacion)` y `evals (legal-core)` de la propuesta de cambio, que el cierre del workflow cuenta; cualquiera en rojo deja el cierre sin aprobar (FR-060, FR-062).
- **SC-002**: Después del run, y fuera de él porque es humano: una persona lee las respuestas del modelo que decide a (b), (d) y (f) en los dos modos y anota en `docs/USO.md` si alguna resume o caracteriza una sentencia cuyo texto no tenía delante. Ninguna release lleva `jurisprudencia` hasta que H25 esté en `main`. Con esa release publicada, en la app de escritorio de Claude con el plugin: «¿existe la STS 1088/2023, de 4 de julio?» se responde sin decir que existe, con la dirección del buscador y las casillas con sus valores; al adjuntar el PDF descargado, con la cita `[ECLI:ES:TS:2023:3144, ROJ: STS 3144/2023]`; y «resúmeme la STS 9999/2023, de 1 de enero» se responde sin resumen. Lo mide una persona: no tiene control en el run.
- **SC-003**: Los once casos del e2e de FR-082 terminan cada uno con su código: 0 en siete —las tres formas y el texto de `preparar`, el ECLI del Tribunal Constitucional y `cotejar` con el documento pedido y con otro— y 2 en cuatro —el ECLI mal formado, `--resolucion` sin `--fecha`, la referencia con `--texto` y el texto sin ficha—. **Control**: los guiones de aceptación fallan en `make ci`.
- **SC-004**: Los cuatro golden de `cita preparar` y los siete de `cita cotejar` coinciden con lo que dan los verbos, salvo `fecha_consulta`, y 11 de 11 salidas son válidas contra su esquema. **Control**: los tests de FR-080 y FR-081 fallan en `make ci`.
- **SC-005**: 0 caminos de dependencia del código del applet `cita` a `internal/httpx`, `net` o `net/http`. **Control**: el test de FR-084 falla en `make ci`.
- **SC-006**: Tras los dos verbos, 0 cambios en la caché y 0 en el grafo del mundo. **Control**: el test de FR-085 falla en `make ci`.
- **SC-007**: Con sesiones sintéticas, una sola respuesta con una cita sin documento cotejado, o con un ECLI que no viene de una operación ni de la pregunta, da veredicto `fallo`; con ninguna, el umbral se cumple; y `umbrales` de `jurisprudencia` tiene 4 elementos. **Control**: los tests de FR-086 fallan en `make ci`.
- **SC-008**: Las preguntas de las evals (d) y (f) contienen los 2 353 bytes de `evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt`, byte a byte, y ese fichero no cambia. **Control**: el test de FR-053 falla en `make ci`.
- **SC-009**: `make ci` termina en verde, con `schema-check` y `skills-check` sin diferencias; `SKILL.md` de `jurisprudencia` tiene menos de 300 líneas; y el servidor anuncia 12 herramientas, las 10 de hoy y las 2 de `cita`. **Control**: `make ci` (`skills-check` falla con 300 líneas o más; el control de conformidad de H21 falla con otro conjunto de herramientas).
- **SC-010**: El README dice, en «¿Y las sentencias?», que kitlegal no consulta el CENDOJ y qué hace; `docs/JURISPRUDENCIA.md` y `CHANGELOG.md` (*Unreleased*) llevan lo de FR-070. Lo comprueba la revisión final contra FR-070.

## Uso, de fuera adentro

Cada salida del hito, desde quien la consume (criterio de uso, ADR 0028). Lo que gana la skill: dar a la persona la consulta exacta sin escribirla de memoria, y saber, antes de citar, qué documento tiene delante y si es el pedido. Volumen de referencia: meses de uso diario, con cientos de sentencias preparadas y cotejadas, la mayoría hace más de una semana. **Ninguna salida crece con ese volumen**: el applet no guarda nada —ni en la caché, ni en el grafo— y cada invocación solo depende de sus argumentos y de su texto (FR-003).

Los tamaños de esta sección son cálculos de la sesión del spec, anteriores a los verbos. Lo medido después con el binario del hito está en contracts/applet-cita.md §10 —una consulta preparada, de 389 a 572 bytes de sobre; un cotejo, de 559 a 1 005; un error, de 294 a 418— y en contracts/evals-jurisprudencia.md §4 y §8 —cada elemento de `umbrales`, de 273 a 388 bytes escrito sin blancos—; donde una cifra de aquí y una de allí difieren, vale la de los contratos.

Cálculos de esta sesión, no medidas: los verbos no existen todavía. Se han compuesto con `jq` cinco `data` de ejemplo con los datos de este spec, con claves supuestas, que fija el plan, y se han contado sus bytes: 276 la consulta por número y fecha, 241 la de un ECLI con su equivalente, 116 la de un texto, 297 los ocho datos de la ficha con su correspondencia y 672 el mismo cotejo con un hallazgo y su explicación. Lo que el sobre añade a `data` son 199 bytes en `territorio resolver Leganés --json`, medido en esta sesión con el binario que había compilado en `bin/` (`v0.3.2-22-g4152c2b`, anterior a la cabeza de `main`). La ficha del fragmento ocupa 316 bytes y el fragmento entero, 2 353. Por herramienta, el sobre va dos veces en el mensaje (H21).

| Salida | Quién la pide, cuántas veces y qué hace con ella | Tamaño | Cuándo deja de darse cada señal |
|---|---|---|---|
| La consulta preparada de una referencia (FR-010, FR-011) | La skill, en el paso de FR-042: una vez por sentencia que la respuesta necesita y no está en la conversación. Copia la dirección y cada casilla con su valor a la respuesta. Una persona, cuando la pide desde la terminal. | ≈ 250 a 300 bytes de `data`, ≈ 500 de sobre; como mucho tres casillas. Una respuesta que necesita cinco sentencias: cinco invocaciones, ≈ 2,5 KB. No crece con lo acumulado. | Cada invocación es independiente y no deja nada. La skill deja de pedirla para una sentencia cuando la persona trae su documento. |
| El equivalente (FR-012) | La skill: puede nombrarlo junto a la referencia, porque viene de una operación (FR-044). | Un dato, del orden de 50 bytes. | Solo se da cuando se deduce: una sentencia del Tribunal Supremo con su ECLI y su ROJ; con cualquier otro órgano o un auto no se da. |
| La dirección de una búsqueda por texto (FR-013) | La skill, en el paso de FR-046: una vez por pregunta por materia. La copia a la respuesta. | ≈ 120 bytes más tres veces, como mucho, los octetos del texto. | Cada invocación es independiente. |
| La declaración de cobertura del Tribunal Constitucional (FR-014) | La persona, desde la terminal, y el agente que pide `cita_preparar` con ese ECLI sin la skill. La skill no la pide: declara no cubierta la sentencia sin preparar nada (FR-047). | Un sobre sin dirección ni casillas, menos de 500 bytes. | Con el candidato `tc` del backlog, que no es de este hito. Se da en cada consulta porque responde a esa consulta. |
| Los datos de la ficha (FR-022) | La skill, en el paso de FR-043: una vez por documento que trae la persona. De ellos sale la cita (FR-044) y lo que la respuesta dice de la sentencia cuando solo tiene la ficha (FR-045). | Ocho datos, ≈ 300 bytes; ≈ 500 de sobre. No crece con el tamaño del documento: del texto solo sale la ficha. | Cada invocación es independiente. |
| La correspondencia entre el ROJ y el ECLI de la ficha (FR-023) | La skill, en cada cotejo (FR-048): solo «no se corresponden» cambia la respuesta —no cita el documento y pide traerlo de nuevo—; con «se corresponden» y con «no se deduce» cita sin mencionarla. | Un dato con tres valores posibles. | Se da en cada cotejo: es un dato del documento, no una señal que se repita sin él. «No se deduce» no llega a la respuesta. |
| Si el documento es el pedido, y el hallazgo (FR-024, FR-025) | La skill, en el paso de FR-043, cuando se había pedido una referencia: con el hallazgo dice qué difiere y no cita el documento como el pedido. | Un hallazgo por invocación, con una o dos diferencias, el cruce si lo hay y su explicación: ≈ 400 bytes. | Deja de darse en el cotejo del documento pedido. No se guarda: no vuelve a aparecer en ninguna otra consulta. |
| Los errores de argumentos (FR-006, FR-015, FR-026) | La skill: corrige la llamada o pide a la persona lo que falta (FR-049). | `{clase, mensaje}`, menos de 1 KB. | En la llamada siguiente bien formada. |
| La cita con su forma (FR-044) | Quien pregunta; una por documento cotejado que la respuesta nombra. | Una línea de unos 70 caracteres. | No es una señal. |
| La línea `⚠ SENTENCIA NO COMPROBADA:` (FR-042) | Quien pregunta; una por sentencia que la respuesta necesita y no está en la conversación. Le dice cuál falta; la consulta preparada, cómo traerla. | Una línea, del orden de 60 a 150 bytes. | Deja de darse cuando la persona trae el documento y se coteja. No se repite en otra respuesta salvo que se vuelva a necesitar esa sentencia sin su documento. |
| La tabla de comandos de `SKILL.md` (FR-040) | El modelo, una vez por conversación en que se activa la skill. | Dos filas, cada una con su orden y su herramienta. `SKILL.md` < 300 líneas. | No da señales. |
| Las dos herramientas en la lista del servidor (FR-030) | El agente, una vez por conexión. | Dos herramientas más, con los esquemas de `--describe` de sus verbos; H21 midió unos 4,6 KB por verbo. No crece con el uso. | No da señales. |
| `umbrales` del informe de `jurisprudencia` (FR-060 a FR-063) | El job, que decide con ellos; el informe final del workflow, que los lee sin modelo; y la persona. Una vez por job. | 4 elementos de ≈ 250 bytes, sobre 18 respuestas por modo: seis evals por tres repeticiones. Tamaño fijo. | Cada job los mide de nuevo sobre su commit. Uno incumplido deja de darse en el primer job que lo cumple. |
| Las tasas del informe (FR-050) | `scripts/workflow/informe.sh`, una vez por run, y la persona. | 24 series: seis evals, dos modos y dos modelos. Lo fija el conjunto de evals, no el uso. | Cada job las mide de nuevo. |

Tiempo: ninguno de los dos verbos espera a nadie. No hay red, ni ritmo, ni reintentos, ni caché que abrir: cada invocación es un proceso local, o una llamada al servidor, que termina con lo que tarda en leer sus argumentos y su texto. No se midió en la sesión del spec, porque los verbos no existían, y el hito no le pone umbral ni control: el plan lo deja sin objetivo («Performance Goals»).

## Fuera de alcance

Del hito, literal:

- «consultar el CENDOJ de forma automática, de cualquier manera —el formulario, la dirección de búsqueda, el documento de una sentencia—, y con ello otro agente, un navegador o resolver un CAPTCHA;»
- «decir que una sentencia existe;»
- «descargar o guardar el texto de una sentencia;»
- «leer un PDF en el binario;»
- «la caché y el grafo;»
- «el filtro por órgano, el número de recurso y los operadores de la búsqueda por texto, sin probar;»
- «el Tribunal Constitucional, que llega por el BOE con el candidato `tc` del backlog;»
- «validar las citas de un texto, que es H8;»
- «el juez con modelo de `jurisprudencia` —su rúbrica, sus casos, su medida y su carpeta—, que es H25;»
- «cambiar la constitución, `scripts/workflow/`, `internal/httpx` o `evidencias/adr-0036/`;»
- «cambiar `boe-legislacion` o `legal-core`;»
- «y la web, que la cambia la persona.»

De lo que el hito no especifica (constitución, «Criterio de decisión autónoma», punto 2):

- Otros verbos del applet `cita`, las citas de normas y `validar` (H8).
- `--fecha` con un ECLI o con `--roj`, y cualquier otro dato en la consulta preparada: el órgano, la jurisdicción, el número de recurso, el ponente o un intervalo de fechas.
- Validar las siglas de un ROJ o el órgano de un ECLI contra una lista; reconocer como del Tribunal Constitucional una referencia que no sea su ECLI; y dar en `data` la dirección del buscador del Tribunal Constitucional.
- Leer una ficha que no sea la del CENDOJ, una ficha a la que le falte alguno de sus ocho datos, o más de un documento por invocación; y dar en `data` cualquier otra cosa del documento que los ocho datos de su ficha.
- Que el binario abra un fichero: una bandera con la ruta del documento. Le llega texto.
- Comparar el órgano de la referencia con el de la ficha, y usar el número de recurso como referencia.
- Detectar un documento falso o alterado: lo asume el ADR 0036.
- Una salida legible propia de los dos verbos: sin `--json`, vale lo que el kernel da a un verbo que no la declara.
- Nodos o aristas de una resolución en el grafo, hallazgos de `graph check` sobre sentencias y cualquier cambio en los verbos de `graph`, de `boe` o de `territorio`.
- Cambiar el applet `mcp` o el paso de empaquetado, y un mecanismo en el binario, en el empaquetado o en la release que excluya `jurisprudencia` hasta H25: que ninguna release la lleve antes lo decide la persona, que es quien etiqueta (ADR 0020).
- Deducir el equivalente entre un ECLI y un ROJ, o su correspondencia, para otro órgano que el Tribunal Supremo, con otras siglas o para un auto (FR-012): pide una fuente versionada, con las siglas y las fichas de esos órganos traídas por una persona, y es de otra entrada.
- Citar con una advertencia un documento cuyo ROJ y cuyo ECLI no se corresponden, y decir algo de la correspondencia cuando `cita cotejar` da «no se deduce» (FR-048); y poner la línea `⚠ SENTENCIA NO COMPROBADA:` a un documento por no corresponderse su ficha.
- Evals de `jurisprudencia` distintas de las seis del hito: una de no activación, una sin binario ni servidor o una con una ficha cuyo ROJ y cuyo ECLI no se corresponden.
- Comprobar en las evals la parte de la cita que va delante del corchete, lo que la respuesta dice que difiere, o si la respuesta dice que una sentencia existe.
- Umbrales del modelo informativo, un umbral de duración para `jurisprudencia` y cualquier umbral nuevo en las otras skills.
- Cambiar `make evals-sondeo`: sus argumentos, su preparación y su salida se quedan como están.
- `references/` nuevas para la skill y cualquier dato nuevo en `data/`.
- `docs/SOURCES.md`, `scripts/verify-sources.sh` y el flujo nocturno: el hito no toca ninguna fuente.
- `CLAUDE.md` y `docs/ROADMAP.md`, que la persona actualiza al fusionar; un ADR nuevo; la release.

## Assumptions

- **«Resume» no se comprueba en este hito** (FR-065). FR-045 es una regla de la skill sin control: lo lee una persona en el cierre (SC-002) y lo decide el juez de H25. El informe final no puede darlo por comprobado.
- **El equivalente y la correspondencia solo se deducen para el Tribunal Supremo** (FR-012, FR-023). Es la única pareja que el repositorio documenta con su ejemplo (`docs/JURISPRUDENCIA.md` §3). Consecuencia, para quien lea el informe final: todo documento que no sea una sentencia del Tribunal Supremo se coteja con «no se deduce», y la skill lo cita sin mencionarlo (FR-048). Rechazada: comparar solo el número y el año para cualquier órgano, que daría por correspondida una ficha con las siglas de un órgano y el ECLI de otro; y escribir de memoria la relación entre siglas y órganos, cuyo equivalente llegaría a `data` y, de ahí, a la respuesta como un identificador que nadie ha leído en un documento.
- **Qué hace la skill con «no se corresponden» no tiene control** (FR-048, FR-065). Con lo decidido en FR-012, solo se da con una ficha de ECLI `TS` y ROJ `STS` cuyo número o año difieren —un documento alterado o una ficha mal copiada—. La eval con una ficha así está fuera de alcance, y `cita_sin_documento` da por leído el ECLI de toda ficha reconocible. Lo lee una persona en el cierre; se anota en `gates/supuestos.md`. Rechazada: citar el documento con una advertencia, que se pierde al copiar la cita a un escrito, y ponerle la línea de FR-042, que es de la sentencia que no está en la conversación.
- **Qué significa «ninguna release lleva `jurisprudencia` hasta H25»**. Se lee como una decisión de la persona sobre cuándo etiqueta (ADR 0020), no como un mecanismo: al fusionar el hito, la skill queda empotrada en el binario de `main`, como toda skill, y las dos herramientas, en su registro. Rechazada: excluirlas del binario o del plugin, que el hito no pide y que contradice que el plugin las lleve «sin tocar su paso».
- **`--fecha` solo con `--resolucion`** (FR-005). El hito define la fecha con el número de resolución y con nada más, y su protocolo dice que una cita sin fecha no se prepara como ROJ. Con un ECLI o con un ROJ no la define, y lo no especificado no se implementa. Rechazada: admitirla y compararla en `cotejar`.
- **La ficha reconocible lleva sus ocho datos** (FR-021). El hito nombra ocho y dice que «un texto sin ficha reconocible es un error de argumentos», sin decir cuántos bastan. Se lee de la forma que menos compromete: con uno de menos no hay ficha, y la sentencia no se cita. Rechazada: aceptar una ficha parcial y dar los datos que haya, que dejaría citar con un documento a medias. No se ha comprobado en esta sesión que todo documento del CENDOJ lleve los ocho: el repositorio solo tiene la ficha de una sentencia del Tribunal Supremo.
- **Lo que cuenta como salida de un `cita cotejar`** (FR-061). El hito dice «cuyo ECLI no está en la salida de ningún `cita cotejar` de la sesión» y, a renglón seguido, que «un ECLI no puede llegar a una respuesta más que desde un documento que se ha cotejado o desde la persona». Si contara cualquier texto de la salida, bastaría dar a `cita cotejar` un ECLI inventado como referencia pedida para que su eco lo dejara citar. Se lee por la segunda frase: el ECLI leído en la ficha. Rechazada: cualquier aparición en la salida.
- **Lo que cuenta como salida de una herramienta** (FR-061). El hito dice «la salida de ninguna herramienta de la sesión». Se lee como la salida de una operación de kitlegal, pedida como orden o como herramienta: en el modo orden no hay herramientas del servidor, y el protocolo de la skill dice «ningún ECLI que no venga de una herramienta o de la persona». Rechazada: contar lo que la sesión lee de cualquier sitio, que dejaría pasar un ECLI copiado del ejemplo de `SKILL.md`. Consecuencia, para quien lea el informe final: un ECLI que el modelo da como argumento a `cita preparar` vuelve en su salida y deja de contar en la segunda condición; en una cita solo vale el que `cita cotejar` leyó en una ficha.
- **El argumento del texto de `cita cotejar` existe también en la orden** (FR-020). El hito dice «por la entrada estándar o en el argumento de la herramienta». Las herramientas salen del `--describe` de su verbo sin tocar el applet `mcp`, así que el argumento es del verbo y la orden lo tiene. Se fija lo que menos sorprende: dado, es el texto. Y en el servidor la entrada estándar es el protocolo: una llamada de herramienta no la lee nunca.
- **La forma del ECLI y la del ROJ** (FR-005) son las del material de lectura que nombra el hito (rama `018-h23-cita-resolver-comprobar`, `internal/core/ids/ecli.go` e `internal/core/ids/roj.go`, leídos en esta sesión). Las dos cotas del ECLI —órgano de hasta 7 caracteres y número de hasta 25— vienen de ahí, donde se escribieron de memoria del estándar europeo que cita `docs/JURISPRUDENCIA.md` §3 (`DOUE-Z-2019-70039`): ese documento no se ha leído en esta sesión, que no tiene red.
- **La codificación del texto de búsqueda** (FR-013). El hito dice «con el texto codificado» y da un ejemplo; `docs/JURISPRUDENCIA.md` §3 recoge dos direcciones probadas a mano, con el espacio como `%20` y la tilde como `%C3%A1`. La regla de FR-013 es la que da esas dos y trata igual cualquier otro carácter. No se ha probado qué hace el buscador con otros caracteres.
- **La dirección del buscador del Tribunal Constitucional** (FR-047) no está escrita en el repositorio: `docs/JURISPRUDENCIA.md` §2 nombra el buscador HJ y sus rutas, no su dirección. Esta sesión no tiene red y no la ha comprobado. El plan la fija en un solo sitio, y queda como supuesto para que la persona la compruebe al leer el informe final.
- **La sentencia del Tribunal Constitucional no lleva la línea ni pasa por `cita preparar`** (FR-047). El hito le da su propio paso, el 7, y a su eval, «la declara no cubierta», sin la línea que sí nombra en (a) y en (b). La línea anuncia una consulta en el buscador del CENDOJ y deja de darse cuando la persona trae el documento y se coteja; el del Constitucional no lleva la ficha del CENDOJ, así que la línea no dejaría de darse nunca. Y nombrada sin su ECLI, `cita preparar` la prepararía como cualquier otra referencia (FR-014): la skill no lo pide de ninguna forma. Rechazada: llevar la línea además de la declaración, y pedir `cita preparar` solo cuando la persona da el ECLI, que son dos caminos para una misma respuesta.
- **Las preguntas de (c) y de (e)** (FR-050). El hito dice «una pregunta por materia» y «una sentencia del Constitucional». La de (c) usa el ejemplo del propio hito, «cláusula suelo»; la de (e), la sentencia 79/2024, que `docs/JURISPRUDENCIA.md` §1 da con su ECLI. La pregunta de (e) lleva el ECLI para que la respuesta pueda nombrarlo sin contar en el umbral: viene de la persona.
- **Cómo se evalúan (c) y (e)** (FR-050). El hito dice que (c) da «la dirección de la búsqueda por texto» y que (e) «la declara no cubierta», y que todo se compara por su forma. Lo que tiene forma en esas dos respuestas es la dirección que el protocolo manda dar y que no haya ninguna cita. En (c) el texto de la búsqueda lo elige el modelo, así que la dirección se compara con la que devolvió la operación de la sesión.
- **La operación pedida en las evals** (FR-050). Que (a), (b) y (c) exijan haber pedido `cita preparar`, y (d) y (f), `cita cotejar`, sale del protocolo: la consulta se da «tomada de `data`» y la cita, de la ficha que leyó `cita cotejar`.
- **`sin_activar` en una skill sin juez** (FR-063). Leído en esta sesión en `internal/evals/umbrales.go`: hoy los umbrales de las respuestas solo se publican para una skill con juez, y `legal-core`, sin juez, tiene `umbrales` vacío. El hito pide que `sin_activar` decida en `jurisprudencia` «como en las otras skills»; se lee como en `boe-legislacion`, la única que lo tiene, y sin dar ninguno a `legal-core`.
- **Las versiones de las skills.** El hito dice «v0». `CHANGELOG.md` nombra la skill por lo que hace para quien la usa; si lleva número, es el primero de su serie.
- **`references/` de la skill.** El hito no da a `jurisprudencia` ningún dato de `data/`: si su carpeta `references/` existe o no lo decide el plan con lo que exija `skills-check`, sin datos nuevos en `data/`.
- **Lo comprobado en esta sesión**: lo leído en el repositorio —la constitución 2.13.0, los ADR 0036 y 0037, `docs/JURISPRUDENCIA.md`, `docs/USO.md`, `evidencias/adr-0036/`, `internal/evals`, `internal/app`, `internal/core/schema`, `internal/arch_test.go`, las dos skills y sus evals, `.github/workflows/evals.yml`, y la rama de lectura— y los tamaños de «Uso, de fuera adentro», calculados con `jq` y `wc -c`, más una ejecución de `kitlegal territorio resolver Leganés --json` con el binario de `bin/` para el tamaño del sobre. No se ha ejecutado ningún test ni `make ci`, ni se ha consultado ninguna fuente.
- **Quedan para el plan**: las claves de `data` y los argumentos de las herramientas; los valores de `fuente` y de `url` del sobre donde FR-004 no los da; el identificador del hallazgo; el paquete del reconocimiento de ECLI y de ROJ y el de la lectura de la ficha; los campos nuevos del formato de eval; la redacción de las preguntas de (d), (e) y (f); y la redacción de `SKILL.md`.
