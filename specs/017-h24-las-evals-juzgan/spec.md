# Feature Specification: H24 · Las evals juzgan el significado con un modelo: «afirma lo que no ha leído» decide, `boe-legislacion` deja de glosar lo que no ha leído, y la lista de expresiones deja de decidir

**Feature Branch**: `017-h24-las-evals-juzgan`

**Created**: 2026-10-05

**Status**: Draft

**Input**: Sección «#### H24 · Las evals juzgan el significado con un modelo: «afirma lo que no ha leído» decide, `boe-legislacion` deja de glosar lo que no ha leído, y la lista de expresiones deja de decidir (adelantado; ADR 0037)» de `docs/ROADMAP.md`, en modo desatendido (ADR 0018).

## Resumen

Quien pregunta a `boe-legislacion` tiene que leer lo que dice la norma, leída con el binario, y nada de un precepto que ninguna herramienta devolvió. Hoy no se cumple, y el job de evals no lo ve.

- **El control mide palabras, no lo que la respuesta dice.** La lista de expresiones (`evals/boe-legislacion/expresiones-prohibidas.yaml`) falla por los dos lados. En el cierre de H7.3 marcó 1 de las 9 respuestas con defecto que encontró la lectura a mano (`docs/USO.md`, 2026-09-30). El 2026-10-04 dejó el job en rojo por dos respuestas correctas de la eval sin binario ni servidor (`docs/USO.md`, 2026-10-04; #116).
- **La skill glosa de memoria lo que no ha leído.** La validación del ADR 0037 lo encontró en la skill de hoy. Lo hace al repetir una remisión: «las entidades del art. 45 (entidades de ámbito territorial inferior al municipio)», cuando el texto leído solo dice «las entidades a que se refiere el artículo 45». Y al declarar lo que no leyó: «No he leído los artículos 67 a 70, que regulan cuándo empieza a contar el plazo…». Por unanimidad del juez son 3 respuestas de 54 en el informe de H7.4; 5 y 6 en el de H21, por modo; y 2 y 4 en el de H22 (ADR 0037, «Validación»). Tres de las glosas de los informes anteriores están mal o se contradicen (`docs/USO.md`, 2026-10-05). Es contenido legal sin fuente (constitución, principio II).
- **Un tercio de las respuestas nombra «el sobre» o un campo de la salida**: «El sobre no trae avisos de vigencia», «rige desde el 2 de octubre de 2016 (`fecha_vigencia` 20161002)». Sale de la prosa de `SKILL.md` v0.1.6, que dice «la respuesta dice lo que trae el sobre» y nombra los dos campos.

El hito no entrega skill nueva. Arregla la que existe (`boe-legislacion` v0.1.7) y cambia el instrumento que la mide: lo que es de significado lo juzga un modelo distinto del que decide, con una pregunta cerrada por clase, la frase citada comprobada sin modelo y tres votos unánimes; y ese juez solo decide si está medido contra casos etiquetados (ADR 0037). Deja además el instrumento con el que H23 medirá «ni resume» una sentencia inventada (constitución, principio VIII).

La entrada humana del hito, que el run no cambia, está en `evidencias/adr-0037/`: la rúbrica (`rubrica.md`), el esquema de la respuesta del juez (`esquema.json`), los 259 casos etiquetados (`casos.yaml`) y la medida versionada (`medida.json`: 212 de 212 defectos marcados con los tres votos y 0 de 47 correctos, con `claude-opus-5-5` y Claude Code 2.1.289).

**Cómo se citan los requisitos de otros hitos.** Los de H5 a H22 se citan sin guion: «H7.4 FR 060», «H5 FR 077». Los ids con guion (FR-001, SC-001…) son siempre de este spec. En este spec, `<modelo>` es el id del modelo que decide (ADR 0031), hoy `claude-sonnet-5-5`, y `<modo>` es `orden` u `herramienta` (H21).

## Clarifications

### Session 2026-10-05

- Q: ¿Qué efecto tiene en el veredicto que el juez deje marcada en `afirma_lo_no_leido` una respuesta de la eval sin binario ni servidor (la 21), y dónde cuenta el tiempo de sus votos? → A: Solo se publica. Se juzga como las demás (mismo voto, mismo mensaje, misma comprobación de la frase, misma regla de tres votos), y sus votos y frases van al informe (FR-060); una marca suya no entra en la `medida` ni en el `total` de ningún umbral, no da motivo de fallo y no cambia si su sesión pasa (la eval sigue decidiendo por su serie y su juicio sin modelo). El tiempo de sus votos no entra en `duracion_del_juez:<modo>`; lo cubre el tope del trabajo (FR-092). `umbrales` se queda en doce elementos, diez que deciden, y «Controles de umbral» en seis filas. Un voto suyo que no llega a darse deja la respuesta «sin juzgar» (FR-007). Lo que queda sin control: una respuesta de esa eval que lleve la línea `⚠ SIN CONSULTA AL BOE:` y además afirme de memoria el contenido del artículo sale en verde, como hoy; el informe la enseña con su frase (auto: conservadora; criterio d; fuente: `docs/ROADMAP.md` H24 «Lo que el juez recibe», «Los umbrales», «La medida del juez, a petición» y Controles; H21 FR 047; ADR 0037, decisiones 3 y 6; 0 de los 259 casos de `evidencias/adr-0037/casos.yaml` son de la eval 21). Rechazada: un umbral propio sin modo que decide, porque el hito no lo define y el juez no se ha medido contra respuestas de esa eval.
- Q: ¿Con qué comprobación decide el job que la frase que cita un voto está en la respuesta: con la tolerancia que tienen hoy las formas fijas del job, o con la de la validación del ADR 0037 (`normal` y `frase_esta` de `evidencias/adr-0037/guiones/juez.py`)? → A: La de la validación. De la frase y de la respuesta se quitan `*`, `_` y el acento grave en cualquier punto; cada serie de blancos, también los saltos de línea, se colapsa en un espacio; se recortan los extremos; y la frase está si, sin quedar vacía, es subcadena de la respuesta. Nada más se tolera. Las piezas de las formas fijas de `internal/evals/avisos.go` no se usan para esto y siguen como están. Medido en esta sesión, sin modelo, sobre los votos versionados y las respuestas de los seis informes: 1.189 votos, 1.061 síes, y ninguno depende de la diferencia entre las dos comprobaciones; la medida de hoy (0 de 212 y 0 de 47) es la misma con cualquiera (auto: criterio c; fuente: `docs/ROADMAP.md` H24 «El voto» y Controles; `evidencias/adr-0037/guiones/juez.py` (`normal`, `frase_esta`, `votar`); `internal/evals/avisos.go`; ADR 0037, decisión 3; constitución, «Criterio de decisión autónoma», punto 1). Rechazada: la tolerancia de las formas fijas, porque daría por no encontrada una copia literal que cruza un salto de línea o pierde un acento grave, y dejaría la comprobación del job distinta de aquella con la que se contó la medida.
- Q: Cuando una persona lanza la ejecución de la medida del juez y la medida versionada no corresponde a la rúbrica, a los casos, al modelo del juez o a su versión de Claude Code que hay, ¿qué hace esa ejecución con la comprobación de FR-041? → A: No la hace. La comprobación de FR-041 que pide FR-043 es de la ejecución normal del job. La ejecución de la medida vota los casos con lo que hay, corresponda o no la medida versionada, e imprime la medida con las cuatro claves de lo que hay y sus dos recuentos; su resultado depende solo de los votos (FR-052, FR-053). No lee la medida versionada para decidir nada ni dice si corresponde: en esa propuesta ya lo dice `make ci` (FR-042) (auto: criterio a; fuente: `docs/ROADMAP.md` H24 «Los umbrales», «La medida del juez», «La medida del juez, a petición»; ADR 0037, decisión 3; `CLAUDE.md`, «Juez con modelo en las evals»; constitución, «Criterio de decisión autónoma», punto 2). Rechazadas: que la haga y decida (deja sin uso la ejecución en el estado para el que existe) y que la haga y solo lo diga (una línea que el hito no define).

## Criterios del hito, literales

Transcripción literal de `docs/ROADMAP.md` §4, H24. El Objetivo y el Alcance se reflejan en los requisitos.

- **Entrega**: «(1) el juez con modelo en el job y en el sondeo, con una pregunta cerrada por clase, la frase citada comprobada sin modelo y tres votos por orden; (2) las clases de cada skill declaradas en datos, con `afirma_lo_no_leido`, que decide, y `cuenta_su_proceso`, que se publica, en `boe-legislacion`; (3) la medida del juez contra los casos etiquetados: versionada, comprobada sin modelo en `make ci` y en el job, y repetida solo a petición de una persona, cuando cambia la rúbrica, los casos, el modelo del juez o su versión de Claude Code; (4) el informe con los votos y las frases de cada respuesta con algún voto afirmativo; (5) la lista de expresiones fuera del veredicto y de los umbrales; (6) `boe-legislacion` v0.1.7, que no dice qué dice ni de qué trata un precepto que no ha leído, y cuya respuesta no nombra «el sobre» ni los campos de la salida; (7) `CHANGELOG.md` (*Unreleased*), con la versión de la skill, y lo que `CONTRIBUTING.md` y `docs/WORKFLOW.md` dicen del job.»
- **Controles**:
  - «`make ci` en verde, con `schema-check` y las reglas del conjunto;»
  - «la sección «Controles de umbral» del plan con una fila por umbral que decide: `afirma_lo_no_leido` en cada modo, los dos de la medida del juez y la duración del juez en cada modo;»
  - «tests del voto con votos grabados: sí con la frase en la respuesta; sí con una frase que no está, nulo y repetido; una frase que solo difiere en blancos y énfasis; y un voto que no llega a darse, «sin juzgar» con su motivo;»
  - «tests de la regla: tres síes, marcada; un no en el primero, sin más votos; sí, sí y no, sin marcar y publicada con sus dos frases; y una clase que solo se publica, con un voto;»
  - «tests del informe con sesiones sintéticas: `afirma_lo_no_leido` con 1 da `fallo` con un motivo que nombra la respuesta y sus frases, y con 0 se cumple; `cuenta_su_proceso` con 3 no cambia el veredicto; y una medida versionada que no corresponde da `fallo` con el motivo del instrumento, sin ninguna sesión abierta;»
  - «tests de la medida versionada: con su rúbrica, sus casos, su modelo del juez y su versión de Claude Code, `make ci` pasa y el job juzga sin medir; con cualquiera de los cuatro cambiado, o con un recuento distinto de 0, `make ci` falla;»
  - «tests de la ejecución de la medida, con votos grabados: un caso etiquetado como defecto sin marcar, o uno correcto marcado, da `fallo` con el caso y sus frases; con los 259 bien, imprime la medida con sus cuatro claves y sus dos recuentos; y no abre ninguna sesión de evals;»
  - «un test de que el juez recibe los textos de las herramientas de la sesión en los dos modos y nada más: ni la skill ni la eval; y otro de que la orden y el mensaje del voto son los de la validación;»
  - «una comprobación en `make ci` de que la rúbrica, el esquema de la respuesta, los casos y la medida de `evals/boe-legislacion/juez/` son idénticos a los de `evidencias/adr-0037/`;»
  - «el control de derivaciones (`TestGrabacionesDerivadas`) con los casos: cada respuesta coincide con la de su informe, byte a byte, y cada defecto derivado solo se diferencia de su sesión en el texto quitado;»
  - «`skills-check` sin drift, y `TestProsaDeLaSkill` con «el sobre» y los dos nombres de campo en el vocabulario que la prosa no usa para lo que la respuesta dice;»
  - «un test de que `Juzgar` ya no marca ninguna sesión por la lista, con las dos respuestas de la eval sin binario ni servidor del 2026-10-04 y las del calibrado de H7.4;»
  - «`TestDefinicionDelJob`: el modelo del juez está fijado por su id completo y es distinto del que decide; la versión de Claude Code de sus votos está fijada aparte de la de las sesiones; y la medida solo se lanza con su etiqueta o con su entrada;»
  - «`CHANGELOG.md` (*Unreleased*); y el job de evals en la propuesta de cambio.»
- **Aceptación**: «en el informe del job de cierre de `boe-legislacion`, los dos umbrales de la medida del juez se publican con los recuentos de la medida versionada y se cumplen, con `decide: true`; `afirma_lo_no_leido:<modelo>:<modo>` lleva `decide: true` y `cumple: true` (0) en los dos modos; `cuenta_su_proceso` se publica con su recuento, junto al de la validación del ADR 0037 (entre 12 y 30 de cada 54); la duración del juez se cumple; `umbrales` no lleva `expresiones_prohibidas` ni `redaccion_no_leida`; ninguna respuesta queda «sin juzgar»; el veredicto es aprobado, el job de `legal-core` sigue aprobado y `red` está vacío; y el informe final da cada umbral que decide como «comprobado por su control». Después, y fuera del run porque es humano: Jorge lanza una vez la medida del juez con su etiqueta sobre la propuesta de cambio, antes de fusionar, porque el código que vota es nuevo y la medida versionada se hizo con los guiones de la validación; si no se cumple, no fusiona, y decide si la rúbrica se corrige fuera de un run o si hace falta un hito de seguimiento. Y lee las respuestas con algún voto afirmativo en cualquiera de las dos clases y las tres de la eval 19 y de la 20 en cada modo, y anota en `docs/USO.md` si el juez acertó.»

Trazabilidad resumida (el detalle, en cada requisito):

| Criterio del hito | Dónde se cumple |
|---|---|
| Entrega (1), el juez en el job y en el sondeo | FR-001 a FR-014, FR-075, FR-076, US2, US7 |
| Entrega (2), las clases en datos | FR-020 a FR-024, FR-030, FR-031 |
| Entrega (3), la medida del juez | FR-032, FR-040 a FR-054, US3, US4 |
| Entrega (4), el informe con votos y frases | FR-060 a FR-062, US5 |
| Entrega (5), la lista fuera del veredicto y de los umbrales | FR-034, FR-070, FR-071, US6 |
| Entrega (6), `boe-legislacion` v0.1.7 | FR-080 a FR-086, US1 |
| Entrega (7), `CHANGELOG.md`, `CONTRIBUTING.md` y `docs/WORKFLOW.md` | FR-087, FR-095 |
| El modelo del juez, su versión y los topes | FR-090 a FR-093 |
| Lo que sustituye | FR-096, «Relación con H5.1 a H22» |
| Controles | FR-100 a FR-113, SC-002 a SC-013 |
| Aceptación | SC-001, SC-014 |

## Relación con H5.1 a H22

`specs/012-h7-2-la-consulta-repetida/`, `specs/013-h7-3-el-umbral-de/` y `specs/014-h7-4-boe-legislacion-sin/` no se editan, ni ningún otro spec, plan o informe de H7.1 a H22: son el registro de sus runs. La decisión de arquitectura ya está tomada en el ADR 0037, aceptado y enmendado antes de lanzar el hito: no hay ADR nuevo (FR-096).

**Sustituye** (lo que dice el hito citado deja de valer, y vale lo de este spec):

- **La comparación «sin juicio de ningún modelo» de las respuestas** (H7.2 a H7.4): que la lista de expresiones juzgue cada sesión, decida por serie (H7.2 FR 054), se publique por sesión y por modelo (H7.2 FR 053) y se aplique por clases y familias (H7.4 FR 030, FR 031, FR 034 y FR 036). Lo sustituyen FR-070 y el juez de FR-001 a FR-014. La clase A de H7.4 (H7.4 FR 010) pasa a ser `cuenta_su_proceso`, que se publica; la clase B (H7.4 FR 011) es un caso de `afirma_lo_no_leido`, que decide.
- **El calibrado de la lista sobre los tres informes** (36, 11 y 9; H7.4 FR 032) y lo que protegía a la lista como juez de respuestas: que ninguna expresión aparezca en los bloques grabados ni en una respuesta hecha de las formas de la skill (H7.4 FR 033). Lo sustituye FR-071: la lista ya no juzga ninguna respuesta, y de su calibrado quedan solo los tres totales, como premisa del test de FR-111. La medida de un juez es ahora la de FR-040 a FR-054.
- **Los umbrales `expresiones_prohibidas` y `redaccion_no_leida`**, del modelo que decide y del informativo, en cada modo (H7.3 FR 002 y FR 004; H7.4 FR 040, FR 042 y FR 044; H21 FR 043, en lo que dice de esos dos). Lo sustituye FR-034. Lo decide la entrada del hito, no un corrector (ADR 0029 y 0037).
- **La viñeta de «Fuera de alcance» de H7.2, H7.3 y H7.4 «juzgar la redacción libre con un modelo o por similitud»** y la «Decisión del mecanismo» de H5.1, en lo que dicen del significado de una respuesta (ADR 0037, punto 1). La similitud sigue descartada, y lo que tiene forma sigue sin modelo (ver «Fuera de alcance»).
- **La salida del sondeo**, en su recuento de respuestas con alguna expresión y su 5 % de referencia (H7.3 FR 065; H7.4 FR 082): FR-075.
- **El tope del trabajo** (H7.3 FR 035; H21 FR 083), en su valor: se recalcula con los votos del juez (FR-092).

**Se queda**: el modelo que decide y su versión de Claude Code (ADR 0031); las repeticiones y la regla por serie (ADR 0016); el contrato de `umbrales` (ADR 0029; H7.3 FR 001 y FR 003); `sin_activar:<modelo>:<modo>` y `duracion_de_las_sesiones:<modo>` (H7.4 FR 041 y FR 043; H21 FR 043); la respuesta a la pregunta y las sesiones que cuentan (H7.4 FR 060, FR 061 y FR 045); los límites de uso y las sesiones sin medir (H7.3 FR 040 a FR 044); una tanda por commit y skill (H7.4 FR 070); los dos modos y la eval sin binario ni servidor, con su juicio por la línea `⚠ SIN CONSULTA AL BOE:` y la ausencia de citas (H21 FR 046); el sondeo, que mide solo el modo orden y no es un veredicto (H7.3 FR 060 a FR 068; H21 FR 050), salvo lo que sustituye FR-075; la comprobación de la prosa de `SKILL.md` (H7.3 FR 091; H7.4 FR 021), que se extiende con FR-085; y H5 FR 077 (`SKILL.md` no nombra evals, el job ni modelos). Todo lo que tiene forma o es un hecho de la sesión —citas, avisos, formas fijas, órdenes, activación, red y duración— se sigue comprobando sin modelo.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Quien pregunta no lee nada de un precepto que nadie leyó (Priority: P1)

Una persona pregunta a `boe-legislacion` por las atribuciones del Pleno. El artículo leído remite a «las entidades a que se refiere el artículo 45». La respuesta traslada la remisión como el texto la da, o lee el art. 45 y lo cita. No añade de memoria de qué trata. Si avisa de que no ha leído un precepto, lo nombra por su número y nada más. Sí puede avisar de que una materia se regula en otra parte de la norma, sin nombrar el precepto ni decir la regla. Y dice qué norma dio la redacción y desde cuándo rige con palabras de quien lee, sin «el sobre» ni nombres de campos.

**Why this priority**: es lo que el hito dice que gana la skill. Una glosa de memoria es contenido legal sin fuente (principio II), y tres de las encontradas eran falsas o se contradecían.

**Independent Test**: en el job de cierre, `afirma_lo_no_leido:<modelo>:<modo>` da 0 en los dos modos (SC-001). En `make ci`, la comprobación de la prosa y `skills-check` (SC-010).

**Acceptance Scenarios**:

1. **Dado** `boe-legislacion` v0.1.7 y el job de cierre con el modelo que decide, **Cuando** el juez vota las respuestas de las evals que activan la skill, **Entonces** `afirma_lo_no_leido:<modelo>:<modo>` tiene `medida` 0, `cumple: true` y `decide: true` en el modo `orden` y en el modo `herramienta` (FR-030, FR-081 a FR-083).
2. **Dado** `SKILL.md` v0.1.6, con las frases «la respuesta dice lo que trae el sobre de `kitlegal boe`» y «qué norma la dio (`norma_modificadora`) y desde cuándo rige (`fecha_vigencia`)», **Cuando** se le aplica la comprobación de la prosa de v0.1.7, **Entonces** falla nombrando esas frases; con `SKILL.md` v0.1.7, pasa (FR-085).
3. **Dado** `SKILL.md` v0.1.7, **Cuando** corre `skills-check`, **Entonces** tiene menos de 300 líneas, la región generada no tiene drift y no nombra evals, el job ni modelos (FR-086).
4. **Dado** el job de cierre, **Cuando** se lee `cuenta_su_proceso:<modelo>:<modo>`, **Entonces** se publica con su recuento y `decide: false` en cada modo, y no cambia el veredicto (FR-031, FR-084).

---

### User Story 2 - El job marca la respuesta que afirma lo que no ha leído, por lo que dice (Priority: P1)

El job de evals pasa cada respuesta del modelo que decide, con la pregunta y los textos que devolvieron las herramientas de su sesión, a un juez con modelo. El juez responde sí o no a cada clase y, si es sí, copia la frase que lo prueba. El job comprueba sin modelo que la frase está en la respuesta. Una respuesta queda marcada en la clase que decide solo con tres síes. Con una marcada, el veredicto es `fallo` y el motivo la nombra con sus frases.

**Why this priority**: sin este control, una glosa de memoria en una de cada diez respuestas sale en verde, como en los cierres de H7.4, H21 y H22.

**Independent Test**: tests del voto, de la regla y del informe con votos grabados, en `make ci` (SC-002 a SC-004, SC-007); el job de cierre (SC-001).

**Acceptance Scenarios**:

1. **Dado** un voto grabado que dice sí con una frase que está en la respuesta, **Cuando** se comprueba, **Entonces** el voto vale y marca (FR-005).
2. **Dado** un voto grabado que dice sí con una frase que no está en la respuesta, **Cuando** se comprueba, **Entonces** es nulo y se pide una vez más; el repetido es el que cuenta (FR-006).
3. **Dado** un voto que dice sí con una frase que solo difiere de la respuesta en blancos y en énfasis de Markdown, **Cuando** se comprueba, **Entonces** la frase está y el voto vale (FR-005).
4. **Dado** un voto que no llega a darse, por un límite de uso o por el tope de tiempo, **Cuando** se escribe el informe, **Entonces** la respuesta queda «sin juzgar» con su motivo y el veredicto es `fallo` (FR-007).
5. **Dado** tres votos que dicen sí, cada uno con su frase comprobada, **Cuando** se aplica la regla, **Entonces** la respuesta queda marcada (FR-010).
6. **Dado** un primer voto que dice no, **Cuando** se aplica la regla, **Entonces** no se pide ningún voto más y la respuesta no queda marcada (FR-010).
7. **Dado** los votos sí, sí y no, **Cuando** se aplica la regla, **Entonces** la respuesta no queda marcada y el informe la publica con sus dos frases (FR-010, FR-060).
8. **Dado** una clase que solo se publica, **Cuando** se aplica la regla, **Entonces** cuenta con el primer voto y no se pide ninguno más por ella (FR-011).
9. **Dado** una sesión del modo `orden` y otra del modo `herramienta`, **Cuando** se compone el mensaje del voto, **Entonces** lleva la pregunta, la respuesta a la pregunta y la salida de cada orden `kitlegal …` o el resultado de cada herramienta del servidor, y no lleva `SKILL.md`, ni lo que la eval espera, ni el resultado de las comprobaciones sin modelo (FR-001, FR-002).
10. **Dado** un informe sintético con 1 respuesta marcada en `afirma_lo_no_leido` en un modo, **Cuando** se escribe, **Entonces** el umbral de ese modo lleva `cumple: false` y el veredicto es `fallo` con un motivo que nombra la respuesta y sus frases; con 0, `cumple: true` y el veredicto no cambia (FR-030, FR-035).

---

### User Story 3 - El juez solo decide si está medido (Priority: P1)

Quien cambia la rúbrica, los casos, el modelo del juez o la versión de Claude Code de sus votos no puede fusionar sin una medida que corresponda: `make ci` falla. El job hace la misma comprobación antes de abrir ninguna sesión y, si la medida no corresponde, termina en rojo diciendo que el instrumento no está medido, sin dar veredicto de la skill. En una ejecución normal no se mide al juez: se publican los recuentos de la medida versionada.

**Why this priority**: un juez sin medir que decide con umbral 0 pondría el job en rojo, o en verde, sin que nadie sepa si acierta.

**Independent Test**: tests de la medida versionada y de las copias, en `make ci` (SC-005, SC-008).

**Acceptance Scenarios**:

1. **Dado** la medida versionada con su rúbrica, sus casos, su modelo del juez y su versión de Claude Code, **Cuando** corre `make ci`, **Entonces** pasa; y el job juzga sin medir al juez y publica los dos umbrales de la medida con los recuentos versionados (FR-041, FR-042, FR-044).
2. **Dado** la rúbrica, los casos, el modelo del juez o su versión de Claude Code cambiados, uno cada vez, **Cuando** corre `make ci`, **Entonces** falla nombrando cuál no corresponde (FR-042).
3. **Dado** una medida con un recuento distinto de 0, **Cuando** corre `make ci`, **Entonces** falla (FR-042).
4. **Dado** una medida versionada que no corresponde, **Cuando** arranca el job, **Entonces** termina sin abrir ninguna sesión, con `fallo` y el motivo de que el instrumento no está medido, y `umbrales` no da por medido ningún umbral del juez (FR-043).
5. **Dado** una copia de `evals/boe-legislacion/juez/` que difiere en un byte de la de `evidencias/adr-0037/`, **Cuando** corre `make ci`, **Entonces** falla nombrando el fichero (FR-023).

---

### User Story 4 - Una persona repite la medida cuando cambia el instrumento (Priority: P2)

Jorge cambia la rúbrica. Pone la etiqueta `evals-medir-juez` en su propuesta de cambio. Esa ejecución no abre sesiones de evals: vota los casos etiquetados, dice si algún defecto queda sin marcar o algún correcto queda marcado, con cada caso y sus frases, e imprime la medida entera en su registro. Jorge la versiona en la misma propuesta, fuera de un run.

**Why this priority**: medir al juez son 683 votos con los casos de hoy, frente a los 108 de juzgar las respuestas; no cabe en una ejecución normal, y sin poder repetirla el instrumento no podría cambiar nunca.

**Independent Test**: tests de la ejecución de la medida con votos grabados y `TestDefinicionDelJob`, en `make ci` (SC-006, SC-012). Con el juez de verdad la lanza una persona (SC-014).

**Acceptance Scenarios**:

1. **Dado** votos grabados en los que un caso etiquetado como defecto no queda marcado, **Cuando** corre la ejecución de la medida, **Entonces** da `fallo` con el caso y sus frases (FR-052).
2. **Dado** votos grabados en los que un caso etiquetado como correcto queda marcado, **Cuando** corre, **Entonces** da `fallo` con el caso y sus frases (FR-052).
3. **Dado** votos grabados con los 259 casos bien, **Cuando** corre, **Entonces** imprime la medida con sus cuatro claves —las huellas de la rúbrica y de los casos, el id del modelo del juez y la versión de Claude Code de sus votos— y sus dos recuentos, 0 de 212 defectos sin marcar y 0 de 47 correctos marcados, y no abre ninguna sesión de evals (FR-051, FR-052).
4. **Dado** esos mismos votos grabados y una medida versionada que no corresponde a lo que hay, **Cuando** corre, **Entonces** vota los 259 casos igual e imprime la medida con las cuatro claves de lo que hay, no las de la versionada, y sus dos recuentos; no termina sin votar ni dice si la versionada corresponde (FR-043, FR-051, FR-052).
5. **Dado** la definición del job, **Cuando** la comprueba `TestDefinicionDelJob`, **Entonces** la medida solo se lanza con la etiqueta `evals-medir-juez` o con su entrada del flujo lanzado a mano, y la etiqueta `evals` no la lanza (FR-050).

---

### User Story 5 - Quien lee el informe ve por qué se marcó cada respuesta (Priority: P2)

Jorge abre el informe del job de cierre. De cada respuesta con algún voto afirmativo, en cualquiera de las dos clases, ve los votos y la frase que cita cada uno, también de las que no llegaron a quedar marcadas. Con eso anota en `docs/USO.md` si el juez acertó.

**Why this priority**: la unanimidad deja pasar los casos dudosos, y `cuenta_su_proceso` no decide: lo que no decide solo sirve si una persona puede leerlo.

**Independent Test**: tests del informe con sesiones sintéticas, en `make ci` (SC-004).

**Acceptance Scenarios**:

1. **Dado** una respuesta con los votos sí, sí y no en `afirma_lo_no_leido`, **Cuando** se escribe el informe, **Entonces** la lleva con sus tres votos y sus dos frases, y dice que no quedó marcada (FR-060).
2. **Dado** 3 respuestas con sí en `cuenta_su_proceso` en un modo, **Cuando** se escribe el informe, **Entonces** `cuenta_su_proceso:<modelo>:<modo>` lleva `medida` 3 y `decide: false`, el veredicto no cambia y el informe lleva cada una con su frase (FR-031, FR-060).
3. **Dado** una respuesta cuyos votos dicen todos no, **Cuando** se escribe el informe, **Entonces** no lleva votos de ella (FR-060).

---

### User Story 6 - La lista de expresiones deja de decidir (Priority: P2)

Una respuesta correcta que ofrece repetir la consulta —«Cuando kitlegal esté disponible, lo consulto y te respondo con el texto y su cita»— ya no hace fallar su sesión. La lista no juzga ninguna respuesta: no está en el juicio de cada sesión, ni en los motivos, ni en el recuento por modelo, ni en `umbrales`, ni en el sondeo. El fichero se queda como el vocabulario que la prosa de `SKILL.md` no usa.

**Why this priority**: es lo que dejó el job en rojo sin motivo el 2026-10-04, y lo que daba 0 de 54 donde el juez encuentra defectos.

**Independent Test**: el test de `Juzgar` sin la lista y los del informe, en `make ci` (SC-004, SC-011).

**Acceptance Scenarios**:

1. **Dado** las dos respuestas de la eval sin binario ni servidor del 2026-10-04 y las respuestas que la lista marcaba en el calibrado de H7.4, **Cuando** las juzga `Juzgar`, **Entonces** ninguna sesión queda marcada ni deja de pasar por una expresión de la lista, y ningún motivo nombra una (FR-070).
2. **Dado** el informe del job de cierre, **Cuando** se lee `umbrales`, **Entonces** no lleva ningún elemento `expresiones_prohibidas:…` ni `redaccion_no_leida:…` (FR-034).

---

### User Story 7 - El sondeo juzga con el juez, sin veredicto (Priority: P3)

Una persona lanza `make evals-sondeo` para ver cómo responde la skill tras un cambio. El sondeo juzga sus respuestas con el juez, con el Claude Code de su equipo, y dice en su salida cuántas quedan marcadas en cada clase y si la medida versionada corresponde a ese Claude Code. No falla por eso: no es un veredicto.

**Why this priority**: es la herramienta de quien cambia la skill; no afecta al veredicto.

**Independent Test**: los tests del sondeo con el `claude` sustituto y votos grabados, en `make ci` (SC-004).

**Acceptance Scenarios**:

1. **Dado** un sondeo con votos grabados, **Cuando** termina, **Entonces** su salida da, en lugar del recuento de expresiones, las respuestas marcadas en `afirma_lo_no_leido` y las que tienen sí en `cuenta_su_proceso`, y sale con 0 sean cuales sean (FR-075).
2. **Dado** un equipo cuyo Claude Code no es el de la medida versionada, **Cuando** se lanza el sondeo, **Entonces** su salida dice que la medida no corresponde a él, juzga igual y sale con 0 (FR-076).

---

### Edge Cases

- **Sí, sí y no**: la respuesta no queda marcada; se publica con sus dos frases (FR-010, FR-060). Es el caso dudoso que la unanimidad deja pasar (ADR 0037, «En contra, y asumido»).
- **Un voto nulo y su repetición, también nula**: en la clase del sí sin frase, el voto no cuenta como sí, así que la respuesta no queda marcada; el informe publica los dos votos como nulos. No es «sin juzgar»: el juez respondió (FR-006).
- **Un voto que no llega a darse en el segundo o el tercero**: la respuesta queda «sin juzgar», no «sin marcar», y el veredicto es `fallo` (FR-007).
- **Una sesión sin textos**: la skill no se activó o no ejecutó ninguna orden. Se juzga igual, y el mensaje del voto lo dice como en la validación (FR-001).
- **Una respuesta de la eval sin binario ni servidor marcada**: se publica con sus frases y no cambia el veredicto ni el umbral de ningún modo (FR-013).
- **Una frase que cruza un salto de línea de la respuesta o difiere de ella en acentos graves**: está, porque la comprobación quita `*`, `_` y el acento grave y colapsa los saltos de línea (FR-005).
- **Medir tras cambiar la rúbrica con la medida versionada sin corresponder**: la ejecución de la medida vota igual e imprime la medida nueva; no comprueba la versionada (FR-043).
- **Una orden que falla**: su salida, con el error, es un texto más de la sesión (FR-001).
- **1 marcada en un modo y 0 en el otro**: el umbral de ese modo no se cumple y el veredicto es `fallo`. Los modos no se suman (H21 FR 043).
- **Una respuesta marcada en `afirma_lo_no_leido` y con sí en `cuenta_su_proceso`**: cuenta en los dos umbrales, una vez en cada uno.
- **Sin respuestas que contar**: si todas las sesiones del modelo que decide quedan sin medir, el total es 0 y la proporción es 0 (contrato del ADR 0029); el veredicto ya es `fallo` por las sesiones sin medir (H7.3 FR 043).
- **Sube la versión de Claude Code de las sesiones** (cambia el alias `sonnet`, ADR 0031): la medida sigue correspondiendo, porque la versión de los votos se fija aparte (FR-091).
- **La sesión de la prueba de red y las del modelo informativo**: no se juzgan (FR-012).
- **Muchas respuestas marcadas**: cada una añade dos votos. Con unos 8 s por voto y uno detrás de otro, las 54 de un modo son unos 432 s, y los 900 s admiten unas 29 marcadas ((900 − 432) / 16). Por encima, `duracion_del_juez:<modo>` no se cumple; con una sola, el veredicto ya es `fallo` (FR-030, FR-033).
- **El aviso de lo que la respuesta no cubre** («el cómputo se regula en otros artículos, que no he leído»): no es un defecto. Lo fija la rúbrica y lo prueban los 47 casos correctos de la medida, no un test nuevo con modelo (FR-022).

## Requirements *(mandatory)*

### Functional Requirements

#### El juez y el voto

- **FR-001**: El juez MUST recibir, de cada respuesta que juzga, tres cosas: la pregunta de la eval; la respuesta que la sesión dio a la pregunta (H7.4 FR 060); y los textos que devolvieron sus herramientas, tal como están en el transcript y en su orden: la salida de cada orden `kitlegal …` en el modo `orden` y el resultado de cada herramienta del servidor en el modo `herramienta`, también cuando la orden o la herramienta falló. Si la sesión no tiene ninguno, se juzga igual y el mensaje lo dice. Comprobable: FR-107.
- **FR-002**: El juez MUST NOT recibir `SKILL.md`, ni lo que la eval espera (sus comandos, sus citas, sus avisos, sus hallazgos), ni el resultado de las comprobaciones sin modelo de esa sesión. Comprobable: FR-107.
- **FR-003**: Cada voto MUST ser una sesión nueva del modelo del juez (FR-090), sin herramientas, sin servidores, sin skills, sin la configuración de ninguna cuenta ni proyecto y sin acceso al repositorio. Responde a cada clase de la rúbrica con sí o no y, si es sí, con la frase de la respuesta que lo prueba. Una sola sesión responde a todas las clases de la skill.
- **FR-004**: La orden con la que se abre el voto, el esquema de su respuesta y el mensaje MUST ser los de la validación del ADR 0037: `voto_real` y `prompt_de` de `evidencias/adr-0037/guiones/juez.py`, y `esquema.json`, con la rúbrica como instrucciones del juez. La medida versionada se hizo con ellos. Comprobable: FR-107.
- **FR-005**: De cada sí, el job MUST comprobar sin modelo que la frase está en la respuesta, con la tolerancia a blancos y a énfasis de Markdown del hito, que es la de `normal` y `frase_esta` de `evidencias/adr-0037/guiones/juez.py`, con las que se calculó el campo `valida` de cada voto de la medida versionada: de la frase y de la respuesta se quitan `*`, `_` y el acento grave en cualquier punto; cada serie de blancos, también los saltos de línea, se colapsa en un espacio; se recortan los extremos; y la frase está si, sin quedar vacía, es subcadena de la respuesta. Nada más se tolera: mayúsculas, acentos y puntuación se comparan tal cual. Los blancos son los de `\s` de Python sobre texto, que incluye los de Unicode (el espacio de no separación, entre ellos); el plan fija cómo lo iguala el código. Las piezas de las formas fijas de `internal/evals/avisos.go` no se usan para esto y siguen como están para los avisos, las citas y los hallazgos. Una frase que solo difiere en eso está. Una frase vacía no está. Comprobable: FR-102.
- **FR-006**: Un voto con algún sí cuya frase no está MUST ser nulo y repetirse una vez, con una sesión nueva; el repetido es el que cuenta, en todas las clases. Si el repetido lleva también un sí cuya frase no está, en esa clase no cuenta como sí. Los votos nulos se publican como nulos (FR-060). Comprobable: FR-102.
- **FR-007**: Un voto que no llega a darse —la sesión del juez no devuelve una respuesta con la forma del esquema: por un límite de uso, por el tope de tiempo del voto o porque la sesión termina con error— MUST dejar la respuesta «sin juzgar». El informe la publica con su motivo, y el veredicto MUST ser `fallo` con un motivo propio, que nombra las respuestas sin juzgar y se distingue sin modelo de los motivos de la skill, como el de una sesión sin medir (H7.3 FR 043). Una respuesta sin juzgar no cuenta como marcada. El tope de tiempo de un voto lo fija el plan. Comprobable: FR-102.

#### La regla

- **FR-010**: En una clase que decide, una respuesta MUST quedar marcada solo si tres votos dicen que sí, cada uno con su frase comprobada. Se vota por orden: un voto por respuesta; el segundo, solo si el primero marcó; el tercero, solo si marcaron los dos. Comprobable: FR-103.
- **FR-011**: Una clase que solo se publica MUST usar el primer voto: la respuesta cuenta en ella si ese voto dice que sí con su frase comprobada. Lo que el segundo y el tercer voto digan de esa clase no cuenta. Comprobable: FR-103.
- **FR-012**: En cada modo, MUST juzgarse las respuestas del modelo que decide en las evals que activan la skill que terminaron con una respuesta a la pregunta: las mismas que forman el `total` de `sin_activar:<modelo>:<modo>` (H7.4 FR 045), 54 por modo con todas terminadas (18 evals, tres repeticiones). MUST NOT juzgarse las del modelo informativo, la sesión de la prueba de red ni las sesiones sin medir o sin terminar. Con ninguna marcada, son 108 votos por ejecución entre los dos modos, y cada respuesta que el primer voto marca añade dos. Los de la eval sin binario ni servidor van aparte (FR-013).
- **FR-013**: Las respuestas del modelo que decide en la eval sin binario ni servidor (la 21), que no es de ningún modo (H21 FR 047) y cuya sesión no tiene textos, MUST juzgarse igual que las demás —el mismo voto, el mismo mensaje, que dice que ninguna herramienta devolvió ningún texto, la misma comprobación de la frase y la misma regla de tres votos (FR-001 a FR-011)— y MUST solo publicarse: sus votos y sus frases van al informe como los de cualquier respuesta con algún voto afirmativo, con si quedó marcada (FR-060). Una marca en esa eval MUST NOT entrar en la `medida` ni en el `total` de ningún umbral, ni dar motivo de fallo, ni cambiar si su sesión pasa: la eval sigue decidiendo por su serie, con su juicio sin modelo (la línea `⚠ SIN CONSULTA AL BOE:` y la ausencia de citas; H21 FR 046 y FR 047), y FR-014 vale también para ella. No está en las 54 respuestas del `total` de cada modo (FR-012), aunque declare `activa: true`, y el tiempo de sus votos MUST NOT entrar en `duracion_del_juez:<modo>` de ningún modo: lo cubre el tope del trabajo, cuyo peor caso cuenta esos votos (FR-092). `umbrales` MUST NOT llevar ningún elemento para esa eval, tampoco con `decide: false`. Un voto suyo que no llega a darse deja la respuesta «sin juzgar» (FR-007). Lo que queda sin control: una respuesta de esa eval con la línea que además afirme de memoria el contenido del artículo sale en verde, como hoy; el informe la enseña con su frase y SC-014 manda leerla. Comprobable: FR-104.
- **FR-014**: El voto del juez MUST NOT cambiar si una sesión pasa en su serie, ni la tasa de ninguna eval, ni la regla por serie (ADR 0016): actúa en el veredicto solo por los umbrales de FR-030 a FR-033 y por las respuestas sin juzgar (FR-007).

#### Las clases, en datos

- **FR-020**: Las clases del juez MUST declararse por skill, bajo `evals/<skill>/juez/`, junto a la rúbrica: cada clase con su nombre, si decide y su umbral. La declaración tiene su esquema en `schemas/`, que se escribe en tareas `[datos]`, y `make ci` la valida contra él: una que no lo cumple es un fichero mal formado, como una eval. Una skill sin `juez/` no tiene juez: ni votos, ni umbrales del juez, ni comprobación de la medida.
- **FR-021**: `boe-legislacion` MUST declarar dos clases: `afirma_lo_no_leido`, que decide, con umbral 0; y `cuenta_su_proceso`, que solo se publica. `legal-core` no declara ninguna.
- **FR-022**: La rúbrica del juez de `boe-legislacion` MUST ser `evidencias/adr-0037/rubrica.md`, sin cambiar una letra. Fija la frontera de las dos clases:
  - `afirma_lo_no_leido`, con la frontera del precepto identificado (ADR 0037). Cuenta que la respuesta exponga una regla que no está en ningún texto que devolvieran las herramientas de la sesión, nombre o no el precepto, o describa una redacción que no devolvieron; que cambie una remisión del texto leído por la descripción de lo remitido; y que diga de qué trata un artículo, un apartado o una disposición que nombra y no leyó, aunque sea con una etiqueta breve. No cuenta que avise de que una materia se regula en otra parte sin nombrar ningún precepto ni decir la regla, que glose un título o un capítulo, que diga qué norma derogó o desarrolla a otra, que identifique una norma (su título, su número, su fecha, su identificador, su rango), ni que parafrasee mal un precepto que sí leyó.
  - `cuenta_su_proceso`: la respuesta dice el estado de una comprobación o lo que el agente tiene, necesita o va a hacer. No lo son las formas fijas que enseña la skill ni «No hay avisos de vigencia sobre este bloque», que es derecho (H7.3).
- **FR-023**: Lo que el job lee MUST ser una copia, bajo `evals/boe-legislacion/juez/`, de cuatro ficheros de `evidencias/adr-0037/`: la rúbrica, el esquema de la respuesta, los casos y la medida. `make ci` MUST fallar, nombrando el fichero, si alguna copia no es idéntica byte a byte a su original. Ningún paso del run escribe en `evidencias/`. Comprobable: FR-108.
- **FR-024**: Los casos etiquetados MUST ser los de `evidencias/adr-0037/casos.yaml`: 259, de los que 145 son defectos conocidos (5 de la bitácora y 140 derivados) y 114 son de la lectura de la validación (67 defectos y 47 avisos correctos). Cada uno nombra el informe versionado y la sesión de donde sale, su etiqueta y su procedencia; un derivado nombra además el bloque cuyo texto se quita. Ninguna respuesta ni ningún texto MUST escribirse a mano: la respuesta se lee del informe, y los textos de sus herramientas se reconstruyen repitiendo las órdenes de la sesión contra las grabaciones, sin modelo y sin red. Comprobable: FR-109.

#### Los umbrales

Cada uno es un elemento de `umbrales` de `informe.json` de `boe-legislacion`, con el contrato del ADR 0029 (H7.3 FR 001). Con los dos modos son doce elementos, diez de ellos con `decide: true`: por modo, los de FR-030, FR-031 y FR-033, `sin_activar` y `duracion_de_las_sesiones`; y los dos de FR-032.

- **FR-030**: `afirma_lo_no_leido:<modelo>:<modo>` MUST ser un umbral nuevo. Su `medida` son las respuestas juzgadas de ese modo (FR-012) que quedan marcadas (FR-010); su `total`, las respuestas juzgadas de ese modo. Lleva `comparacion` `"<="`, `umbral` 0, `decide: true` y una `descripcion` que nombra la medida. Es 0 porque el principio II no admite una proporción de contenido legal sin fuente. **Control**: con una respuesta marcada, el veredicto es `fallo` y el job de evals de `boe-legislacion` sale en rojo (FR-035).
- **FR-031**: `cuenta_su_proceso:<modelo>:<modo>` MUST ser un elemento nuevo con `decide: false`: su `medida` son las respuestas juzgadas de ese modo con sí en esa clase (FR-011), y su `total`, las juzgadas de ese modo. Se publica y MUST NOT cambiar el veredicto. No es un requisito con umbral: el hito lo deja publicado, y que decida queda fuera de alcance por decisión de la persona (ADR 0037, punto 7). El contrato pide un `umbral`: lleva 0, el de la clase que decide, y su `cumple` solo informa.
- **FR-032**: La medida del juez MUST publicarse en dos umbrales por clase que decide, con `decide: true` y `comparacion` `"<="` con `umbral` 0: los casos etiquetados como defecto que no quedan marcados, sobre los casos etiquetados como defecto; y los etiquetados como correctos que quedan marcados, sobre los etiquetados como correctos. Una ejecución normal del job los publica con los recuentos de la medida versionada: hoy, 0 de 212 y 0 de 47. En este spec se nombran `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` y `medida_del_juez:afirma_lo_no_leido:correctos_marcados`; el plan puede fijar otros nombres. **Control**: con un recuento distinto de 0, `make ci` falla (FR-042) y el job termina con `fallo` sin abrir ninguna sesión (FR-043).
- **FR-033**: `duracion_del_juez:<modo>` MUST ser un umbral nuevo: los segundos que tardan los votos de las respuestas de la tanda de ese modo, desde que empieza el primero hasta que termina el último, sin `total`, con `comparacion` `"<="`, `umbral` 900 y `decide: true`. Va aparte de `duracion_de_las_sesiones:<modo>`, que sigue con sus 900 s, y no cuenta la ejecución de la medida (FR-051). Son 900 porque la validación dio unos 8 s por voto (ADR 0037): las 54 respuestas de un modo, sin ninguna marcada, son unos 432 s con un voto detrás de otro. **Control**: por encima de 900 s, el veredicto es `fallo`, con un motivo de la ejecución y no de la skill, como el de la duración de las sesiones (H7.3 FR 051), y el job sale en rojo (FR-035).
- **FR-034**: `sin_activar:<modelo>:<modo>` y `duracion_de_las_sesiones:<modo>` MUST seguir como están. `umbrales` MUST NOT llevar ningún elemento `expresiones_prohibidas:…` ni `redaccion_no_leida:…`, de ningún modelo ni modo. Lo decide la entrada del hito, no un corrector (ADR 0029 y 0037).
- **FR-035**: Un umbral con `decide: true` que no se cumple MUST poner el veredicto en `fallo`, con un motivo que nombra el umbral y su medida, y el job MUST salir en rojo (H7.3 FR 003). El motivo de `afirma_lo_no_leido:<modelo>:<modo>` nombra además cada respuesta marcada, por su sesión, con sus tres frases. Uno que se cumple MUST NOT cambiar el veredicto. Comprobable: FR-104.
- **FR-036**: Ninguna corrección del run MUST cumplir un umbral de FR-030, FR-032 o FR-033 rebajándolo, dejándolo en `decide: false`, sacando respuestas o casos de su total, ni cambiando la rúbrica, los casos o la medida (ADR 0029; FR-045). Si el cierre da una marca que parece errónea, queda en el informe final para que la lea la persona.
- **FR-037**: `legal-core` MUST seguir sin umbrales: su `umbrales` es `[]` (H7.3 FR 006), y su job no abre el juez.

#### La medida versionada

- **FR-040**: La medida versionada es `evidencias/adr-0037/medida.json`, con su copia (FR-023): la clase, la fecha, el id del modelo del juez, la versión de Claude Code de sus votos, las huellas de la rúbrica y de los casos, y los dos recuentos. La primera es la de la validación del ADR 0037 y ya está en `main`: el hito MUST NOT repetirla ni escribirla.
- **FR-041**: Una medida **corresponde** si sus cuatro claves son las de lo que hay: la huella de la rúbrica y la de los casos son las de las copias que el job lee; el id del modelo del juez es el fijado (FR-090); y la versión de Claude Code es la fijada para sus votos (FR-091). **Se cumple** si sus dos recuentos son 0: ningún defecto sin marcar y ningún correcto marcado. La comprobación no usa ningún modelo.
- **FR-042**: `make ci` MUST fallar si la medida versionada no corresponde, diciendo cuál de las cuatro no coincide, o si alguno de sus dos recuentos no es 0. Así, una propuesta que cambia la rúbrica, los casos, el modelo del juez o su versión de Claude Code no pasa `make ci` hasta que lleva su medida. Comprobable: FR-105.
- **FR-043**: La ejecución normal del job de una skill con juez MUST hacer la comprobación de FR-041 antes de abrir ninguna sesión, de evals o del juez. La ejecución de la medida (FR-050) MUST NOT hacerla ni decir si la medida versionada corresponde (FR-051). Si la medida no corresponde o no se cumple, MUST terminar sin abrir ninguna, con el veredicto `fallo` y un motivo que dice que el instrumento no está medido y cuál de las cuatro no coincide; `umbrales` MUST NOT llevar ningún umbral del juez (los de FR-030 a FR-033); y el job sale en rojo. No da veredicto de la skill. Comprobable: FR-104.
- **FR-044**: Una ejecución normal del job MUST NOT medir al juez: no vota ningún caso etiquetado. Publica los dos umbrales de FR-032 con los recuentos de la medida versionada.
- **FR-045**: El run MUST NOT cambiar la rúbrica, los casos ni la medida, ni sus copias una vez escritas idénticas, y MUST NOT lanzar la ejecución de la medida. Fija el modelo del juez y la versión de Claude Code de sus votos que dice la entrada (FR-090, FR-091).

#### La ejecución de la medida, a petición

- **FR-050**: La medida del juez MUST lanzarse solo de dos formas: con la etiqueta `evals-medir-juez` en una propuesta de cambio, o con una entrada del flujo lanzado a mano, como la prueba de red. La etiqueta `evals`, la apertura de una propuesta y el despacho sin esa entrada MUST NOT lanzarla, y el run solo pone la etiqueta `evals`. Comprobable: FR-112.
- **FR-051**: Esa ejecución MUST NOT abrir ninguna sesión de evals. Por cada clase que decide, vota sus casos etiquetados con la regla de FR-005 a FR-010, con el modelo del juez y la versión de Claude Code fijados, y con los textos reconstruidos de cada caso (FR-024). Con los casos de hoy son 683 votos si la medida se cumple: tres por cada uno de los 212 defectos y uno por cada uno de los 47 correctos. Vota corresponda o no la medida versionada a lo que hay: no la lee para decidir nada ni dice si corresponde (en esa propuesta lo dice `make ci`, FR-042), que es el estado de US4 y de FR-054.
- **FR-052**: La ejecución MUST imprimir la medida entera en su registro, como el job imprime `informe.json`: con las claves de FR-040, que son las de lo que hay y no las de la medida versionada, y sus dos recuentos, lista para que una persona la versione. MUST dar `fallo`, y salir en rojo, si un caso etiquetado como defecto no queda marcado o uno etiquetado como correcto queda marcado, nombrando cada caso con sus frases. Comprobable: FR-106.
- **FR-053**: Si un voto de la ejecución de la medida no llega a darse (FR-007), el caso queda sin juzgar: la ejecución MUST dar `fallo` con ese motivo y los casos sin juzgar, y MUST NOT imprimir una medida. Una medida con casos sin juzgar no es una medida.
- **FR-054**: La medida la versiona una persona, fuera de un run, en la misma propuesta que cambia la rúbrica, los casos, el modelo del juez o su versión de Claude Code. Ningún paso del run ni del job MUST escribirla en el repositorio. Una medida que no se cumple no se versiona: si alguien lo hace, `make ci` falla (FR-042).

#### El informe

- **FR-060**: `informe.json` MUST llevar, de cada respuesta juzgada con algún voto afirmativo en cualquier clase —también un sí nulo, y también si no llegó a quedar marcada—: su sesión; por clase, sus votos en orden, cada uno tal como lo dio el juez para esa clase (los campos de `esquema.json`) y con si su frase está en la respuesta; y si quedó marcada. De una respuesta sin ningún voto afirmativo no lleva votos. El plan fija las claves y lo que de ello escribe `informe.md`. Comprobable: FR-104.
- **FR-061**: El informe MUST llevar cada respuesta «sin juzgar» con su sesión y su motivo (FR-007).
- **FR-062**: El informe MUST NOT llevar el recuento de respuestas con alguna expresión por modelo, ni las expresiones de cada sesión, ni ningún motivo de sesión por una expresión (FR-070).

#### La lista de expresiones

- **FR-070**: La lista de expresiones MUST dejar de juzgar respuestas: sale de `Juzgar` —ninguna sesión deja de pasar por una expresión—, de los motivos de cada sesión, del recuento por modelo, de `umbrales` (FR-034) y del sondeo (FR-075). Comprobable: FR-111.
- **FR-071**: `evals/boe-legislacion/expresiones-prohibidas.yaml` MUST quedarse, con su esquema y su validación en `make ci`, como el vocabulario que la prosa de `SKILL.md` no usa (H7.3 FR 091; H7.4 FR 021; FR-085). Dejan de exigirse las comprobaciones que comparan la lista con respuestas o protegen ese juicio: el calibrado sobre los tres informes (36, 11 y 9; H7.4 FR 032) y la comparación con los bloques grabados y con las formas de la skill (H7.4 FR 033). De ese calibrado quedan solo sus tres totales, como premisa del test de FR-111.

#### El sondeo

- **FR-075**: `make evals-sondeo` MUST juzgar con el juez (FR-001 a FR-011) las respuestas de su modelo en las evals que activan la skill, con el modelo del juez fijado (FR-090) y el Claude Code del equipo de quien lo lanza. Su salida MUST dar, en lugar del recuento de expresiones y su 5 %, las respuestas marcadas en cada clase que decide y las que tienen sí en cada clase que se publica, sobre las juzgadas, y las respuestas sin juzgar con su motivo. Sigue publicando solo lo agregado, sin ser un veredicto y saliendo con 0 sean cuales sean sus recuentos (H7.3 FR 065 y FR 066). Lo demás del sondeo no cambia: mide solo el modo orden (H21 FR 050) y ningún paso del workflow lo lanza (H7.3 FR 068).
- **FR-076**: La salida del sondeo MUST decir si la medida versionada corresponde al Claude Code con el que ha votado, y MUST NOT fallar ni dejar de juzgar si no corresponde.

#### `boe-legislacion` v0.1.7

- **FR-080**: El research MUST trazar cada cambio de `SKILL.md` a su causa, con las respuestas marcadas en la validación del ADR 0037 (`evidencias/adr-0037/`: los casos, la lectura y los votos), como en H7.2, H7.3 y H7.4: qué enseña `SKILL.md` v0.1.6, línea a línea, que lleva a glosar un precepto no leído y a nombrar «el sobre» y los campos. Una prohibición más, sin su causa, no satisface este requisito. Comprobable: el research da la causa de cada defecto con esas respuestas y el plan traza a ella cada cambio.
- **FR-081**: La respuesta MUST NOT decir qué dice ni de qué trata un artículo, un apartado o una disposición que ninguna herramienta devolvió en esta pregunta. Vía: una pregunta sobre un artículo que remite a otro (las evals 03, 08, 10 y 13 a 17) o cuya consulta falla; diferencia: una frase sobre un precepto sin cita ni lectura, que en la validación resultó falsa en tres respuestas. Lo mide `afirma_lo_no_leido:<modelo>:<modo>` (FR-030).
- **FR-082**: Una remisión del texto leído MUST trasladarse como el texto la da, sin cambiarla por la descripción de lo remitido. Si lo que dice el precepto remitido importa para responder, se lee, y entonces se cita.
- **FR-083**: Cuando la respuesta dice que no ha leído o que no ha podido leer un precepto, MUST nombrarlo por su número y nada más. Sí MUST poder avisar de lo que no cubre: que una materia se regula en otra parte de la norma, sin nombrar el precepto ni decir la regla.
- **FR-084**: La respuesta MUST decir qué norma dio la redacción leída y desde cuándo rige con palabras de quien lee, sin nombrar «el sobre» ni los campos de la salida (`fecha_vigencia`, `norma_modificadora`). Lo publica `cuenta_su_proceso:<modelo>:<modo>` (FR-031), que no decide.
- **FR-085**: Fuera de las órdenes y de la región generada, la prosa de `SKILL.md` MUST NOT usar «el sobre» ni los nombres `fecha_vigencia` y `norma_modificadora` para decir lo que la respuesta dice o lleva. Esas tres expresiones entran en el vocabulario que la prosa no usa, y la comprobación de la prosa de `make ci` MUST fallar con ellas en ese uso: como mínimo, con las dos frases de v0.1.6 que el hito señala, «la respuesta dice lo que trae el sobre de `kitlegal boe`: sus avisos y, de la redacción leída, qué norma la dio (`norma_modificadora`) y desde cuándo rige (`fecha_vigencia`)» (paso 5) y «qué norma le dio esa redacción (`norma_modificadora`) y desde cuándo rige (`fecha_vigencia`)» («Redacción modificada»). El plan fija cómo delimita la comprobación ese uso. Lo que cambie bajo `schemas/` va en tareas `[datos]`. Comprobable: FR-110.
- **FR-086**: MUST mantenerse todo lo demás de v0.1.6: la forma de la cita y de los avisos, la comprobación en la misma orden que la lectura, la línea `⚠ SIN CONSULTA AL BOE:` y lo que la skill dice de la redacción anterior. `SKILL.md` MUST seguir por debajo de 300 líneas (hoy tiene 298), con frontmatter válido, la región generada intacta (`make skills-check` sin drift) y sin nombrar evals, el job ni modelos (H5 FR 077). **Control**: `skills-check` falla en `make ci` con 300 líneas o más.
- **FR-087**: `CHANGELOG.md` (*Unreleased*) MUST registrar `boe-legislacion` v0.1.7 con lo que cambia para quien la usa: no dice qué dice ni de qué trata un precepto que no ha leído, traslada las remisiones como el texto las da y dice la vigencia sin «el sobre» ni nombres de campos.

#### El modelo del juez, su versión y los topes

- **FR-090**: El modelo del juez MUST ser `claude-opus-5-5`, fijado por su id completo en `.github/workflows/evals.yml` junto a `MODELO_DE_EVALS`, y MUST ser distinto del modelo que decide. Comprobable: FR-112.
- **FR-091**: La versión de Claude Code de los votos del juez MUST ser 2.1.289, la de la medida, fijada en el mismo fichero y aparte de `VERSION_DE_CLAUDE_CODE`, que es la de las sesiones: subir una no obliga a repetir la medida de la otra. Comprobable: FR-112.
- **FR-092**: El tope del trabajo de evals (`timeout-minutes`) MUST cubrir el peor caso con los votos del juez, y el de la ejecución de la medida, el suyo. `TestDefinicionDelJob` los recalcula con las evals y los casos del repositorio (H7.3 FR 035). **Control**: `TestDefinicionDelJob` falla en `make ci` si un tope no cubre su peor caso.
- **FR-093**: Dentro del run, ninguna sesión MUST abrir el juez (ADR 0032): los tests usan votos grabados y no abren ninguna sesión con modelo; el juez de verdad juzga en el job de la propuesta de cambio.

#### Documentación y relación con otros hitos

- **FR-095**: `CONTRIBUTING.md` y `docs/WORKFLOW.md` MUST decir del job lo que cambia: el juez y sus clases, la regla de los votos, los umbrales nuevos y los que salen, la medida versionada y qué hace fallar `make ci`, cómo se lanza y se versiona la medida (la etiqueta `evals-medir-juez`, una persona, fuera de un run), el modelo del juez y su versión de Claude Code, lo que da ahora el sondeo, y que la lista es solo el vocabulario de la prosa. `CHANGELOG.md` (*Unreleased*) MUST registrarlo junto a FR-087.
- **FR-096**: Ningún spec, plan ni informe de H7.1 a H22 MUST editarse. Este spec nombra lo que sustituye («Relación con H5.1 a H22»). No hay ADR nuevo, y `scripts/workflow/` MUST NOT cambiar.

#### Controles y Definition of Done

- **FR-100**: `make ci` MUST quedar en verde, con `schema-check` y las reglas del conjunto.
- **FR-101**: La sección «Controles de umbral» del plan MUST tener una fila por umbral que decide de este hito, seis: `afirma_lo_no_leido:<modelo>:<modo>` en cada modo (FR-030), los dos de la medida del juez (FR-032) y `duracion_del_juez:<modo>` en cada modo (FR-033), cada una en `evals:boe-legislacion:<nombre>`. `cuenta_su_proceso` no tiene fila.
- **FR-102**: Los tests del voto, con votos grabados, MUST cubrir: sí con la frase en la respuesta; sí con una frase que no está, nulo y repetido; una frase que solo difiere en blancos y énfasis —al menos en un salto de línea y en un acento grave, además de en `*` o `_`, que es lo que separa esta comprobación de la de las formas fijas (FR-005)—; y un voto que no llega a darse, «sin juzgar» con su motivo.
- **FR-103**: Los tests de la regla MUST cubrir: tres síes, marcada; un no en el primero, sin más votos; sí, sí y no, sin marcar y publicada con sus dos frases; y una clase que solo se publica, con un voto.
- **FR-104**: Los tests del informe con sesiones sintéticas MUST cubrir: `afirma_lo_no_leido` con 1, que da `fallo` con un motivo que nombra la respuesta y sus frases, y con 0, que se cumple; `cuenta_su_proceso` con 3, que no cambia el veredicto; una medida versionada que no corresponde, que da `fallo` con el motivo del instrumento, sin ninguna sesión abierta; y una respuesta de la eval sin binario ni servidor marcada con tres síes, que va al informe con sus votos y sus frases (FR-060) y, frente a las mismas sesiones sin esa marca, no cambia la `medida` ni el `total` de ningún umbral, no añade ningún elemento a `umbrales`, no da motivo de fallo y no cambia si su sesión pasa (FR-013).
- **FR-105**: Los tests de la medida versionada MUST cubrir: con su rúbrica, sus casos, su modelo del juez y su versión de Claude Code, `make ci` pasa y el job juzga sin medir; con cualquiera de los cuatro cambiado, o con un recuento distinto de 0, `make ci` falla.
- **FR-106**: Los tests de la ejecución de la medida, con votos grabados, MUST cubrir: un caso etiquetado como defecto sin marcar, o uno correcto marcado, da `fallo` con el caso y sus frases; con los 259 bien, imprime la medida con sus cuatro claves y sus dos recuentos; y no abre ninguna sesión de evals. Estos tests MUST NOT depender de que la medida versionada corresponda, y el caso de los 259 bien MUST ejercitarse al menos una vez con una medida versionada que no corresponde: vota e imprime la medida con las cuatro claves de lo que hay, no las de la versionada (FR-051, FR-052). Con una que corresponde, el test no distinguiría esta lectura de la que termina sin votar.
- **FR-107**: Un test MUST comprobar que el juez recibe los textos de las herramientas de la sesión en los dos modos y nada más: ni la skill ni la eval. Otro MUST comprobar que la orden y el mensaje del voto son los de la validación (FR-004).
- **FR-108**: Una comprobación en `make ci` MUST fallar si la rúbrica, el esquema de la respuesta, los casos o la medida de `evals/boe-legislacion/juez/` no son idénticos a los de `evidencias/adr-0037/`.
- **FR-109**: El control de derivaciones (`TestGrabacionesDerivadas`) MUST cubrir los casos: cada respuesta coincide con la de su informe, byte a byte, y cada defecto derivado solo se diferencia de su sesión en el texto quitado.
- **FR-110**: `skills-check` MUST quedar sin drift, y `TestProsaDeLaSkill` MUST llevar «el sobre» y los dos nombres de campo en el vocabulario que la prosa no usa para lo que la respuesta dice (FR-085).
- **FR-111**: Un test MUST comprobar que `Juzgar` ya no marca ninguna sesión por la lista, con las dos respuestas de la eval sin binario ni servidor del 2026-10-04 y las del calibrado de H7.4: las 36, 11 y 9 respuestas de los informes de H7.1, H7.2 y H7.3 que la lista del repositorio marca, aplicada a la respuesta de cada sesión de esos informes. Esos tres totales son la premisa del test, que identifica sus respuestas y le impide pasar en vacío; el reparto por eval y por familia de H7.4 FR 032 no se exige (FR-071).
- **FR-112**: `TestDefinicionDelJob` MUST fallar si el modelo del juez no está fijado por su id completo o es el que decide; si la versión de Claude Code de sus votos no está fijada aparte de la de las sesiones; o si la medida se puede lanzar sin su etiqueta o sin su entrada.
- **FR-113**: `CHANGELOG.md` (*Unreleased*) MUST llevar lo de FR-087 y FR-095, y el job de evals MUST ejecutarse en la propuesta de cambio (lo hace el workflow tras la revisión final); su informe da lo que pide SC-001.

### Key Entities

- **Clase**: una pregunta cerrada de la rúbrica, que se responde con sí o no. Decide o solo se publica.
- **Voto**: una sesión nueva del modelo del juez sobre una respuesta, que responde a todas las clases y, en cada sí, cita una frase.
- **Voto nulo**: un voto con un sí cuya frase no está en la respuesta. Se repite una vez.
- **Respuesta marcada**: en una clase que decide, la que tiene tres síes con su frase comprobada.
- **Respuesta sin juzgar**: aquella de la que un voto no llegó a darse.
- **Rúbrica**: las instrucciones del juez de una skill, con sus clases y la frontera de cada una.
- **Caso etiquetado**: una respuesta de un informe versionado, con su etiqueta (defecto o correcto) y su procedencia; un derivado, además, con el bloque cuyo texto se quita.
- **Medida del juez**: los dos recuentos del juez contra los casos etiquetados, con las cuatro claves de las que depende.
- **Ejecución de la medida**: la del job que vota los casos y no abre sesiones de evals. La lanza una persona.
- **Umbral del informe**: un elemento de `umbrales` con el contrato del ADR 0029.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: En el informe del job de cierre de `boe-legislacion`:
  - los dos umbrales de la medida del juez se publican con los recuentos de la medida versionada y se cumplen, con `decide: true`;
  - `afirma_lo_no_leido:<modelo>:<modo>` lleva `decide: true` y `cumple: true` (0) en los dos modos;
  - `cuenta_su_proceso` se publica con su recuento, junto al de la validación del ADR 0037 (entre 12 y 30 de cada 54);
  - la duración del juez se cumple (≤ 900 s en cada modo);
  - `umbrales` no lleva `expresiones_prohibidas` ni `redaccion_no_leida`;
  - ninguna respuesta queda «sin juzgar».

  Además, el veredicto es aprobado, el job de `legal-core` sigue aprobado y `red` está vacío; y el informe final da cada umbral que decide como «comprobado por su control».

  **Control**: el job de evals de `boe-legislacion` sale en rojo, con el veredicto `fallo`, en cualquiera de estos casos: una respuesta queda marcada en `afirma_lo_no_leido` en algún modo (FR-030); los votos de un modo pasan de 900 s (FR-033); la medida versionada no corresponde o no se cumple (FR-043); alguna respuesta queda sin juzgar (FR-007); un umbral que sigue (`sin_activar`, `duracion_de_las_sesiones`) no se cumple; alguna sesión queda sin medir; o una serie que decide no pasa. El cierre del workflow cuenta ese rojo.
- **SC-002**: Los tests del voto distinguen los cuatro casos de FR-102: de los cuatro, 1 vale con su frase, 1 es nulo y se repite exactamente 1 vez, 1 vale con una frase que solo difiere en blancos y énfasis, y 1 deja la respuesta «sin juzgar». **Control**: fallan en `make ci`.
- **SC-003**: Los tests de la regla distinguen los cuatro casos de FR-103: 3 votos y marcada; 1 voto y sin marcar; 3 votos, sin marcar y 2 frases publicadas; y 1 voto para la clase que solo se publica. **Control**: fallan en `make ci`.
- **SC-004**: Los tests del informe distinguen los casos de FR-104: `afirma_lo_no_leido` con 1 da `fallo` y con 0 se cumple; `cuenta_su_proceso` con 3 deja el veredicto igual; con una medida que no corresponde hay `fallo` y 0 sesiones abiertas; y 1 respuesta de la eval sin binario ni servidor marcada con tres síes va al informe con sus 3 votos y sus 3 frases y deja los mismos elementos de `umbrales`, con la misma `medida` y el mismo `total`, 0 motivos de fallo por ella y su sesión como estaba. **Control**: fallan en `make ci`.
- **SC-005**: Con la medida versionada, `make ci` pasa; con cada una de sus 4 claves cambiada, o con cada uno de sus 2 recuentos distinto de 0, falla: 6 de 6 mutaciones en rojo. **Control**: `make ci` (FR-042, FR-105).
- **SC-006**: La ejecución de la medida con votos grabados da `fallo` con 1 defecto sin marcar y con 1 correcto marcado; con los 259 casos bien, imprime la medida con 0 de 212 y 0 de 47, y al menos 1 vez lo hace con una medida versionada que no corresponde: vota los 259 e imprime la medida con las 4 claves de lo que hay, no las de la versionada; y abre 0 sesiones de evals. Ninguno de esos tests depende de que la medida versionada corresponda. **Control**: sus tests fallan en `make ci` (FR-106).
- **SC-007**: El mensaje del voto lleva los textos de las herramientas en los 2 modos y 0 bytes de `SKILL.md`, de la eval o de las comprobaciones sin modelo; y la orden y el mensaje son los de la validación. **Control**: los tests de FR-107 fallan en `make ci`.
- **SC-008**: Las 4 copias de `evals/boe-legislacion/juez/` son idénticas a las de `evidencias/adr-0037/`: 4 de 4. **Control**: la comprobación de FR-108 falla en `make ci`.
- **SC-009**: En los 259 casos, la respuesta coincide byte a byte con la de su informe, y cada uno de los 140 derivados solo se diferencia de su sesión en el texto quitado. **Control**: `TestGrabacionesDerivadas` falla en `make ci` (FR-109).
- **SC-010**: `SKILL.md` de `boe-legislacion` tiene menos de 300 líneas, sin drift en la región generada, y 0 usos de «el sobre», `fecha_vigencia` o `norma_modificadora` para lo que la respuesta dice. **Control**: `skills-check` y la comprobación de la prosa fallan en `make ci` (FR-086, FR-110).
- **SC-011**: `Juzgar` marca 0 sesiones por la lista, con las dos respuestas del 2026-10-04 y las 56 del calibrado de H7.4 (36, 11 y 9; FR-111). **Control**: el test de FR-111 falla en `make ci`.
- **SC-012**: La definición del job tiene el modelo del juez fijado por su id completo y distinto del que decide, 2 versiones de Claude Code fijadas por separado, y 0 formas de lanzar la medida que no sean su etiqueta o su entrada; y cada tope cubre su peor caso. **Control**: `TestDefinicionDelJob` falla en `make ci` (FR-092, FR-112).
- **SC-013**: `make ci` en verde, con `schema-check` y las reglas del conjunto, y `CHANGELOG.md` (*Unreleased*) con `boe-legislacion` v0.1.7. **Control**: `make ci`.
- **SC-014**: Después del run, y fuera de él porque es humano: Jorge lanza una vez la medida del juez con su etiqueta sobre la propuesta de cambio, antes de fusionar, porque el código que vota es nuevo y la medida versionada se hizo con los guiones de la validación; si no se cumple, no fusiona, y decide si la rúbrica se corrige fuera de un run o si hace falta un hito de seguimiento. Y lee las respuestas con algún voto afirmativo en cualquiera de las dos clases y las tres de la eval 19 y de la 20 en cada modo, y anota en `docs/USO.md` si el juez acertó. Lo mide una persona: no tiene control en el run.

## Uso, de fuera adentro

Cada salida del hito, desde quien la consume (criterio de uso, ADR 0028). El hito no cambia el binario: lo que la skill lee de `kitlegal boe` y de `kitlegal graph check` sigue acotado como en H7.1, por la norma y los bloques pedidos. Nada de lo que entregan el job, la ejecución de la medida o el sondeo crece con el uso del kit: cada ejecución parte de las evals y de los casos del repositorio, no de lo consultado en meses de uso (cientos de normas y miles de bloques).

| Salida | Quién la pide, cuántas veces y qué hace con ella | Tamaño | Cuándo deja de darse cada señal |
|---|---|---|---|
| La respuesta de `boe-legislacion` v0.1.7 | La persona que pregunta; una por pregunta. Lee el texto citado y sus avisos (FR-081 a FR-084). | Lo que ocupa la norma leída, con 0 frases sobre preceptos no leídos. El aviso de lo que no cubre es una frase de unos 150 bytes, como mucho una por materia que el texto leído deja fuera. No cuenta lo acumulado: con miles de bloques consultados es la misma. | El aviso sale mientras la respuesta no lea esa materia; si la lee, lo sustituye la cita. Que no ha podido leer un precepto sale solo en la pregunta en que la consulta falla. |
| El cuerpo de `SKILL.md` v0.1.7 | El modelo, una vez por conversación en que se activa la skill. | Menos de 300 líneas (hoy 298; FR-086). | No da señales. |
| `umbrales` del informe de `boe-legislacion` | El job, que decide con ellos el veredicto (FR-035); el informe final del workflow, que los lee sin modelo (ADR 0029); y la persona. Una vez por job. | 12 elementos de unos 300 bytes, unos 3,6 KB; `[]` en `legal-core`. Fijo: no crece con el uso. | Cada job los mide de nuevo sobre su commit. Los dos de la medida repiten los recuentos versionados hasta que una persona versiona otra medida. |
| Los votos y las frases del informe (FR-060) | La persona que lee el informe: Jorge en la aceptación (SC-014) y quien mire un job en rojo. Una vez por job. | Acotado por las respuestas juzgadas (108 entre los dos modos, FR-012, y las 3 de la eval sin binario ni servidor, FR-013), no por el uso. Un voto son un motivo de una o dos frases, una frase de 300 caracteres como mucho y, en la clase que decide, el precepto: unos 1,2 KB con las dos clases. Con `cuenta_su_proceso` entre 12 y 30 de cada 54, son de 24 a 60 respuestas con un voto: de 30 a 70 KB. El máximo, con todas marcadas y cada voto repetido por nulo, 111 × 6 × 1,2 KB ≈ 800 KB. | Cada job los escribe de nuevo. Una respuesta sin votos afirmativos no deja ninguno. |
| El motivo de `fallo` por una respuesta marcada (FR-035) | La persona, y la reparación del cierre del workflow, que lee los motivos. Uno por umbral incumplido. | El umbral, su medida y, por respuesta marcada, su sesión y tres frases de 300 caracteres como mucho: alrededor de 1 KB por respuesta marcada. Con la skill corregida, ninguno. | Deja de darse en el primer job sin respuestas marcadas en ese modo. |
| El motivo «el instrumento no está medido» (FR-042, FR-043) | Quien cambia la rúbrica, los casos, el modelo del juez o su versión: lo lee en `make ci` y en el job. | Una línea, menos de 300 bytes, que dice cuál de las cuatro no coincide. | Deja de darse cuando una persona versiona una medida que corresponde y se cumple. |
| La medida impresa por su ejecución (FR-052) | La persona que la lanzó, una vez por lanzamiento: la copia a `evidencias/` y a su copia. | Unos 700 bytes (la de hoy tiene 717). Con `fallo`, además, cada caso con sus frases: alrededor de 1 KB por caso, y como mucho los 259. | Una por lanzamiento. No se repite sola. |
| Las líneas del juez en la salida del sondeo (FR-075, FR-076) | Quien lanza el sondeo, una vez por lanzamiento. | Una línea por clase (2) y una sobre la medida: menos de 500 bytes, más una línea por respuesta sin juzgar. | Una por lanzamiento. |
| La declaración de clases de `evals/<skill>/juez/` (FR-020) | El job y `make ci`, una vez por ejecución. | 2 clases en `boe-legislacion`; fija. | No da señales. |

## Fuera de alcance

Del hito, literal:

- «juzgar con el modelo nada que tenga forma o sea un hecho de la sesión: citas, avisos, formas fijas, órdenes, activación, red y duración se siguen comprobando sin modelo (H5.1, *Decisión del mecanismo*, vigente para todo eso);»
- «la fidelidad de la respuesta al precepto que sí leyó, que es un candidato del backlog;»
- «cambiar el binario o `legal-core`, y en `boe-legislacion`, nada distinto de lo que dice el alcance;»
- «que `cuenta_su_proceso` decida; extender el juez a `legal-core`; la clase de `jurisprudencia`, que la declara H23;»
- «cambiar el modelo que decide, las repeticiones o la regla por serie (ADR 0016 y 0031);»
- «cambiar la rúbrica, los casos, la medida, el modelo del juez o su versión de Claude Code: llegan validados, y si el cierre da una marca que parece errónea, queda en el informe final para que la lea la persona;»
- «medir al juez en una ejecución normal del job o dentro del run;»
- «`scripts/workflow/`, y editar el spec, el plan o los informes de H7.1 a H22.»

De lo que el hito no especifica (constitución, «Criterio de decisión autónoma», punto 2):

- Juzgar con el juez las respuestas del modelo informativo o la sesión de la prueba de red (FR-012).
- Que el voto del juez cambie si una sesión pasa en su serie, o la tasa de una eval (FR-014).
- Que una marca del juez en la eval sin binario ni servidor decida, con un umbral propio o sin él (FR-013): el juez no se ha medido contra respuestas de esa eval (ninguno de los 259 casos es suyo). Si la persona lo quiere, el camino está fuera del run: casos etiquetados de esa eval en `evidencias/adr-0037/casos.yaml`, la medida repetida y el umbral escrito en la entrada del hito.
- Que la ejecución de la medida compruebe, o diga, si la medida versionada corresponde a lo que hay (FR-043, FR-051).
- Medir la similitud entre la respuesta y los textos, o cualquier otro juicio con modelo que no sea una pregunta cerrada de la rúbrica.
- Una clase nueva, o cambiar la frontera de las dos que hay.
- Evals nuevas, o cambiar las preguntas, los comandos o el juicio sin modelo de las que hay.
- Que el sondeo liste los votos o las frases de cada respuesta, escriba un informe o un veredicto, mida el modo herramienta o lance la medida del juez (FR-075).
- Que el job o el run escriban la medida en el repositorio, o que una medida se versione sola (FR-054).
- Meter en los casos una marca que una persona da por errónea: lo hace la persona, fuera de un run, y el arreglo va a la rúbrica.
- Grabar respuestas nuevas del BOE: los casos se reconstruyen con las grabaciones que ya hay.
- Reintentar o esperar un voto más allá de la repetición del nulo (FR-006): un voto que no llega a darse deja la respuesta sin juzgar (FR-007).
- Un ADR nuevo, y la release.

## Assumptions

- **El voto nulo es del voto entero.** El hito dice «un sí con una frase que no está es un voto nulo, y se repite una vez», y un voto es una sesión que responde a las dos clases (`esquema.json` las exige). Se sigue lo que hizo la validación (`nulo` y `juzgar_caso` de `guiones/juez.py`): un sí sin su frase en cualquier clase anula el voto y el repetido sustituye al primero en las dos. Un repetido que sigue sin su frase no marca, por la regla del hito: «solo si tres votos dicen que sí, cada uno con su frase comprobada».
- **Un voto que no llega a darse incluye la sesión del juez que termina con error o sin la forma del esquema.** El hito nombra el límite de uso y el tope de tiempo; tratar cualquier otra falta de respuesta como un «no» daría por limpia una respuesta que nadie juzgó.
- **Una ejecución de la medida con casos sin juzgar no imprime una medida** (FR-053). El hito dice qué hace con un defecto sin marcar o un correcto marcado, y no con un voto que no llega a darse; con 683 votos, un límite de uso es una vía real. Imprimir recuentos sobre los casos que sí se votaron daría por medida una que no lo es.
- **`duracion_del_juez:<modo>` es tiempo de reloj**, del primer voto al último de las respuestas de ese modo, como `duracion_de_las_sesiones:<modo>` lo es de su tanda. El hito lo calcula «con un voto detrás de otro»; si el plan vota varios a la vez, la medida baja, no sube.
- **El `umbral` de `cuenta_su_proceso` es 0.** El hito no le da valor y el contrato del ADR 0029 lo pide. Con `decide: false` no cambia nada del veredicto.
- **«Junto al de la validación del ADR 0037 (entre 12 y 30 de cada 54)» es la referencia con la que la persona lee el recuento**, como los «frente a» de la Aceptación de H7.4. El informe del job no lleva esa cifra.
- **Los nombres de los dos umbrales de la medida** son de este spec y cumplen el patrón del contrato (`^[a-z0-9_.:-]+$`). El hito no los nombra. Si el plan fija otros, sustituyen a los de FR-032 y FR-101.
- **La salida del sondeo sigue siendo agregada** (H7.3 FR 065): el hito dice que el sondeo juzga y qué dice de la medida, y no que liste votos ni frases.
- **`scripts/workflow/informe.sh` no necesita cambios.** Lee el recuento de expresiones del informe con un valor por omisión cuando falta, y los `umbrales` por su nombre (leído en el repositorio, no ejecutado en esta sesión).
- **`TestProsaDeLaSkill` es el nombre que da el hito.** Hoy fija la extracción de la prosa, y quien lee `SKILL.md` es la subprueba `prosa-de-la-skill` de `TestEvalsDelRepositorio` (`internal/evals/conjunto_test.go`). Esa extracción no cuenta lo escrito como código en línea, que es como v0.1.6 escribe los dos nombres de campo: cómo alcanza la comprobación ese uso lo fija el plan (FR-085).
- **Qué comprobaciones de la lista se retiran** (FR-071) sale de lo que el hito dice que se sustituye: «la comparación sin juicio de ningún modelo de las respuestas» y su calibrado. Se quedan las que miran la prosa de `SKILL.md` y el formato del fichero.
- **Los 259 casos se reconstruyen con las grabaciones del repositorio.** La validación lo hizo así con las 423 respuestas de los seis informes (`evidencias/adr-0037/LEEME.md`, «Cómo se hizo»). El hito no graba nada.
- **Lo comprobado en esta sesión**, sin ejecutar ningún test ni `make ci`: `casos.yaml` tiene 259 casos, 212 con la etiqueta `defecto` y 47 con `correcto`, y 5, 140 y 114 por procedencia; `medida.json` dice 0 de 212 y 0 de 47 con `claude-opus-5-5` y 2.1.289; `SKILL.md` tiene 298 líneas; y `.github/workflows/evals.yml` fija `claude-sonnet-5-5` con 2.1.284 y un tope de 240 minutos. Las demás cifras (3 de 54; 5 y 6; 2 y 4; entre 12 y 30 de cada 54; 8 s por voto) son del hito y del ADR 0037.
- **Los nombres técnicos que aparecen** (`Juzgar`, `TestGrabacionesDerivadas`, `TestProsaDeLaSkill`, `TestDefinicionDelJob`, `voto_real`, `prompt_de`, `MODELO_DE_EVALS`, `VERSION_DE_CLAUDE_CODE`, `evals-medir-juez`, `informe.json`) existen en el repositorio o los fija el hito. Quedan para el plan: el formato de la declaración de clases y su esquema; las claves de los votos en el informe; los nombres de las variables del juez en la definición del job; el tope de tiempo de un voto y cuántos se piden a la vez; cómo se graban los votos de los tests; cómo delimita la comprobación de la prosa el vocabulario nuevo; y el texto de las líneas del sondeo.
