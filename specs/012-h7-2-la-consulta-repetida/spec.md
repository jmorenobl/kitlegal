# Feature Specification: H7.2 · La consulta repetida con evidencia coherente, y respuestas sin la maquinaria interna

**Feature Branch**: `012-h7-2-la-consulta-repetida`

**Created**: 2026-09-29

**Status**: Draft

**Input**: Sección «#### H7.2 · La consulta repetida con evidencia coherente, y respuestas sin la maquinaria interna» de `docs/ROADMAP.md`, en modo desatendido (ADR 0018).

## Resumen

Las evals del cierre de H7.1 (informe del job sobre `a8aeb0d`, `specs/011-h7-1-graph-check-acotado/gates/evals/boe-legislacion.json`) destaparon dos defectos que `boe-legislacion` arrastra desde H7 (bitácora `docs/USO.md`, entrada del 2026-09-29):

1. **La eval de la consulta repetida es imposible.** Su grafo previo siembra una redacción «anterior» del art. 21 LPAC hecha a mano (vigencia 20151002, anterior a la entrada en vigor de la ley), mientras la respuesta del BOE que sirve la caché trae una sola redacción (20161002). Una de las tres sesiones de Sonnet 5 no traslada el cambio y remata con «Este texto coincide con lo que te habría confirmado antes», algo que la skill no puede saber.
2. **Las respuestas narran la comprobación.** 34 de las 51 respuestas de Sonnet 5 en evals que activan la skill (67 %) —y 0 de las 30 de Haiku 4.5— llevan ruido interno en el primer párrafo («Sin hallazgos en la memoria de consultas…», «Código 0 sin hallazgos…»), pese a que `SKILL.md` lo prohíbe desde H7. Ninguna eval lo mide.

H7.2 no entrega skill nueva: arregla lo que H7 y H7.1 dieron a la que existe (constitución, principio VIII), antes de H20 y de cualquier release que lleve H7. Son cuatro piezas:

1. **La eval de la consulta repetida sobre un artículo modificado de verdad**: el art. 118 LCSP (`BOE-A-2017-12902`, bloque `a1-30`), con su redacción original, derivada de la respuesta grabada, en el grafo, y la vigente en la caché.
2. **`boe-legislacion` v0.1.2**: la respuesta no menciona la maquinaria interna ni afirma lo dicho en otra conversación, con el cambio de `SKILL.md` trazado a la causa de raíz.
3. **Una lista de expresiones prohibidas en la respuesta**, en el formato común de eval, que se comprueba sin modelo en todas las evals de la skill que la activan y decide en las que deciden.
4. **Un escenario de quickstart** que reproduce, con órdenes que cualquiera repite, la aceptación manual de H7.1.

**Cómo se citan los requisitos de otros hitos.** Los requisitos, historias y criterios de H5, H7 y H7.1 se citan sin guion —«H7 FR 085», «H7.1 FR 050», «H7.1 SC 007», «H7.1 US1.1», «H5 FR 077»—, para no confundirlos con los de este spec. Los ids con guion (FR-001, SC-001…) son siempre de este spec.

## Clarifications

### Session 2026-09-29

- Q: ¿Cómo se abren las dos conversaciones de Claude Code del escenario del quickstart (FR-060, FR-061) y cómo llega cada respuesta a la comprobación mecánica? → A: Cada conversación es una invocación no interactiva de Claude Code: `claude -p` con la pregunta literal de la eval nueva, sin `--continue` ni `--resume`, de modo que la segunda es una conversación nueva, lanzada desde el directorio de trabajo del escenario y con `KITLEGAL_CACHE_DIR` en el directorio temporal preparado. Cada invocación escribe su salida en un fichero propio de ese directorio temporal, y la comprobación mecánica (FR-061) lee de él la respuesta final de la sesión, la misma que juzgaría el job. La persona solo ejecuta las órdenes del quickstart: no escribe ni copia nada a mano. El plan fija las banderas exactas (modelo, los permisos justos para que la sesión ejecute `kitlegal`, formato de salida y de dónde se toma la respuesta final). (auto: criterio a; fuente: docs/ROADMAP.md, H7.2, Alcance, «Escenario de quickstart»: «El quickstart fija las órdenes y una comprobación mecánica de las dos salidas […] con órdenes que cualquiera repite»; spec FR-060 («Sus órdenes MUST poder repetirse tal cual»), FR-061 y US5; scripts/evals.sh (las sesiones del job son `claude -p` con la pregunta de la eval))
- Q: ¿Dónde instala el escenario del quickstart el binario y la skill del hito para que Claude Code los use en las dos conversaciones? → A: Solo dentro del directorio temporal del escenario. El binario se construye desde la rama del hito en ese directorio y va primero en el PATH de las órdenes del escenario, de modo que la skill invoca ese `kitlegal` desde el PATH, como lo invocan las skills (ADR 0019). La skill se instala con ese mismo binario, `kitlegal skills install --host claude`, en ámbito local, en el directorio de trabajo de las conversaciones (`.agents/skills/` y el enlace relativo en `.claude/skills/`). Nada cambia fuera del directorio temporal: ni la instalación de kitlegal de la persona, ni las skills de su cuenta, ni el árbol, el índice o el historial de git (FR-060); tampoco quedan las conversaciones guardadas fuera de él (el job ya las abre sin persistencia de sesión). Como el escenario tiene que ejercer «el binario y la skill del hito» (FR-060), el plan fija cómo se garantiza que la conversación carga la skill instalada en ese directorio y no otra con el mismo nombre que la persona tenga en su cuenta (por ejemplo, la de una release anterior). (auto: criterio c; fuente: spec FR-060, US5 y Assumptions; CLAUDE.md, «Decisiones ya tomadas», Distribución (ADR 0019): `kitlegal skills install` instala en local por defecto y enlaza `--host claude` en `.claude/skills/` con un enlace relativo, las skills invocan `kitlegal` desde el PATH y «`make install` es solo el bucle de desarrollo»)
- Q: ¿Con qué se implementa la comprobación mecánica de las dos respuestas del quickstart (FR-061), que compara la lista «como en FR-051»? → A: Un punto de entrada de prueba de `internal/evals` con la etiqueta de compilación `evals`, invocado desde el quickstart con `go test -tags evals -run '^…$' ./internal/evals/ -args …`, como `TestPrepararSesion`, que el escenario ya usa para preparar el directorio con la misma preparación que el job (FR-060). Lee las dos respuestas guardadas y comprueba las tres condiciones de FR-061 con el mismo código que el job: la lista de la skill con su comparación de FR-051, y la comparación por la forma de `⚠ REDACCIÓN MODIFICADA:` que existe desde H7.1, más las dos fechas en la misma línea; dice cuál de las tres condiciones falla, si alguna. No añade criterios a `Juzgar`, no cambia el formato común de eval ni el informe del job, y queda fuera de `make ci` y del run (FR-062). Aplica la lista y su comparación a dos respuestas, que es lo que FR-061 pide («comparada como en FR-051»), así que cabe en el límite de FR-056 (la lista, FR-050 a FR-055). El plan fija su nombre y sus banderas. (auto: criterio c; fuente: spec FR-051, FR-056, FR-060, FR-061 y FR-062; internal/evals/job_test.go (TestPlanDeSesiones, TestPrepararSesion y TestInformeDelJob: puntos de entrada con la etiqueta `evals` invocados con `go test … -args`, fuera de `make ci`); constitución, «Criterio de decisión autónoma», punto 1 (ni atajos ni ñapas))
- Q: ¿Se aplica la lista de expresiones prohibidas a la sesión de la prueba de red de `boe-legislacion` (la que el job añade con la etiqueta `evals-prueba-de-red` o la entrada `prueba_de_red`, y cuya pregunta pide decir qué devolvieron dos órdenes que fallan) y entra en el recuento por modelo de FR-053? → A: Se aplica y se publica, pero no entra en el recuento por modelo. La sesión de la prueba de red se juzga con la eval 01, que activa la skill, así que `Juzgar` le aplica la lista como a cualquier sesión (FR-052) y el informe publica sus expresiones encontradas, como las de cada sesión (FR-053; Aceptación del hito: «cada sesión publica sus expresiones prohibidas»). El recuento por modelo frente al umbral (FR-053, SC-001) se hace sobre las sesiones de las series que pide el plan —con el conjunto actual, 51 respuestas de Sonnet 5 y 30 de Haiku 4.5—, sin la de la prueba de red, que el informe ya trata como una serie no planificada que no decide. Si su respuesta lleva expresiones de la lista —la pregunta ampliada pide decir qué devolvieron dos órdenes—, esa sesión no pasa y sus expresiones quedan a la vista junto a su tasa, sin decidir; lo que de ella decide el veredicto sigue siendo solo que `red` esté vacío. La ejecución de cierre no la incluye: `scripts/workflow/cierre.sh` pone solo la etiqueta `evals`. (auto: criterio c; fuente: spec FR-052, FR-053, FR-054 y SC-001; docs/ROADMAP.md, H7.2, «Umbral» (2 de 51 con Sonnet 5 y 1 de 30 con Haiku 4.5 con el conjunto actual); specs/006-h5-skill-boe-legislacion/contracts/job-de-evals.md §6, enmienda del ADR 0016 (la sesión se juzga con la eval 01, forma una serie que el plan no pide, no decide y «no mide la calidad de la skill sino la garantía de red»); internal/evals/plan.go (PruebaDeRed) e informe.go (series no planificadas con pregunta ampliada); .github/workflows/evals.yml (etiqueta `evals-prueba-de-red` y entrada `prueba_de_red`, la vía real) y scripts/workflow/cierre.sh)

## Criterios del hito, literales

Transcripción literal de `docs/ROADMAP.md` §4, H7.2, de lo que este spec tiene que cumplir tal cual; el único cambio es el id de H7.1 de la Entrega, escrito sin guion por la regla anterior. El Objetivo y el Alcance se reflejan en cada requisito.

- **Entrega**: «(1) la eval de la consulta repetida sobre un artículo modificado de verdad, con su redacción anterior real en el grafo y la vigente en la caché, en lugar de la eval 19; (2) `boe-legislacion` v0.1.2, que no menciona la maquinaria interna ni afirma lo dicho en otra conversación, con la causa de raíz corregida en `SKILL.md`; (3) una lista de expresiones prohibidas en la respuesta, en el formato común de eval, que se comprueba sin modelo en todas las evals de la skill que la activan y decide en las que deciden; (4) un escenario de quickstart que reproduce la aceptación manual de H7.1 (H7.1 SC 007).»
- **Controles**:
  - «`make ci` en verde, con `skills-check` y `schema-check` sin drift y con las reglas del conjunto;»
  - «el control de derivaciones (`TestGrabacionesDerivadas`) exige más al grafo previo de una eval: servida en lugar de la grabación, `boe` da exactamente una de las redacciones que trae la grabada (fecha de vigencia, norma modificadora, texto y huella), y una derivada con una fecha, un texto o una huella que la grabada no trae no pasa;»
  - «tests de formato, en los que una lista mal formada es un fichero mal formado;»
  - «tests de `Juzgar`: una respuesta sin expresiones prohibidas; con una de cada familia; con las variantes toleradas; y con la forma `⚠ REDACCIÓN MODIFICADA:`, que no es una expresión prohibida;»
  - «una comprobación de calibrado: sobre las 93 respuestas del informe de H7.1, la lista marca exactamente las 35 que cuenta, eval por eval, la entrada de `docs/USO.md` del 2026-09-29 (las 34 con ruido interno y la de «te habría confirmado antes»), y ninguna otra;»
  - «ninguna expresión de la lista aparece en el texto de los bloques grabados para las evals de `boe-legislacion`, para que transcribir un artículo no dé un falso positivo;»
  - «`CHANGELOG.md` (*Unreleased*);»
  - «y el job de evals en la propuesta de cambio.»
- **Aceptación**: «en el informe del job de cierre, la eval nueva declara `⚠ REDACCIÓN MODIFICADA:` y su tasa, cada sesión publica sus expresiones prohibidas y el recuento de cada modelo queda dentro del umbral (≤ 5 % de sus respuestas en las evals que activan la skill, frente al 67 % de Sonnet 5 en H7.1), con el veredicto aprobado (el hito cambia `SKILL.md`, Definition of Done §1.10) y `red` vacío; y el escenario del quickstart da la forma en la primera conversación y no en la segunda, sin expresiones prohibidas en ninguna.»
- **Fuera de alcance**: «cambiar el binario (`graph check`, sus clases, su salida) o el protocolo de la skill más allá de lo que piden el Alcance y la regla 7; tocar `internal/evals` y `schemas/eval.yaml.json` más allá de la lista de expresiones prohibidas y de la derivación real del grafo previo; promover a decisoria ninguna eval informativa; extender la lista a `legal-core` o a otras skills; juzgar la redacción libre de la respuesta con un modelo o por similitud (H5.1, *Decisión del mecanismo*), incluidas las paráfrasis que la lista no recoge; la misma pregunta dos veces dentro de una misma conversación: la segunda puede apoyarse en el texto ya leído, y la memoria de consultas sirve a la pregunta hecha en otra conversación; editar el spec, el plan, los guiones o las grabaciones de H7 y H7.1 fuera de lo que retira el Alcance.»

Trazabilidad resumida (el detalle, en cada requisito):

| Criterio del hito | Dónde se cumple |
|---|---|
| Entrega (1), la eval nueva y su grafo previo | FR-001 a FR-004, FR-010 a FR-014, FR-020, FR-021, US2 |
| Entrega (2), `boe-legislacion` v0.1.2 | FR-040 a FR-048, US1, US4 |
| Entrega (3), la lista y su decisión | FR-050 a FR-056, US3 |
| Entrega (4), el quickstart | FR-060 a FR-062, US5, SC-002 |
| Qué evals deciden | FR-030 |
| Relación con H7 y H7.1 | FR-070 |
| Controles | FR-080 a FR-087, SC-003 a SC-008 |
| Aceptación | SC-001, SC-002 |

## Relación con H7 y H7.1

`specs/010-h7-internal-graph-grafo/` y `specs/011-h7-1-graph-check-acotado/` no se editan: son el registro de sus runs. No cambia ninguna decisión de arquitectura, así que no hay ADR nuevo (FR-070).

**Sustituye** (lo que dicen H7 y H7.1 deja de valer y vale lo de este spec):

- H7.1 FR 050 y H7.1 US1.1 (la eval 19 sobre el art. 21 LPAC, con la redacción 20151002 en el grafo y la 20161002 en la caché): FR-001 a FR-003 y US2.
- H7.1 FR 053, y en H7 FR 085 lo que dice del estado previo (una «versión anterior del bloque» cualquiera): FR-002 y FR-010 a FR-012. El resto de H7 FR 085 —una eval de consulta repetida que se prepara sin red— se cumple con la eval nueva.
- H7 FR 095, en lo que dice de «el estado previo de la eval» como fixture derivada: FR-010 (sigue entrando por una tarea `[datos]`).
- H7.1 FR 055 y H7 SC 012 (la eval informativa de la consulta repetida y su tasa en el informe): FR-004 y FR-030, ahora sobre la eval nueva.
- H7.1 SC 006: SC-001.
- H7.1 SC 007 (la misma pregunta dos veces en Claude Code, comprobada por la persona): FR-060 a FR-062 y SC-002, con órdenes y una comprobación mecánica.
- H7 FR 082 y H7.1 FR 046, en lo que piden decir cuando `graph check` sale con otro código («que no se ha podido comprobar la memoria de consultas»), y H7.1 US1.5: FR-043. Lo demás de H7 FR 082 (código 0 es un resultado) se queda.
- H7.1 FR 044 (con 0 y sin `version-obsoleta`, no decir nada de la memoria): lo amplían FR-040 y FR-041.
- H7.1 FR 042, en las fechas de la forma fija: FR-042 precisa que van tal como las da el hallazgo.
- H7.1 FR 047 (`SKILL.md` no cambia más allá de lo de H7.1): FR-047.
- Los supuestos de H7 y H7.1 de que el estado previo de la eval se deriva de la grabación cambiando fecha de vigencia y texto: FR-010. Las derivadas del e2e siguen como estaban (FR-013).

Todo lo demás de H7 y H7.1 sigue en vigor.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Quien pregunta lee la norma, no la herramienta (Priority: P1)

Una persona pregunta a `boe-legislacion` por un artículo. La respuesta le da el texto citado, sus avisos de vigencia y, solo si la redacción cambió desde la consulta anterior, la línea `⚠ REDACCIÓN MODIFICADA:`. No le cuenta cómo se comprobó: ni la memoria de consultas, ni `graph check`, ni códigos de salida, ni «hallazgos», ni las clases del binario, ni el JSON, ni el sobre. Lo que no se pudo consultar se lo dice por lo que significa para ella.

**Why this priority**: es el defecto medido que la release llevaría a dos de cada tres respuestas con Sonnet 5, y lo que el hito dice que gana la skill.

**Independent Test**: en el job de evals, la lista de expresiones prohibidas sobre todas las respuestas de las evals que activan la skill (SC-001).

**Acceptance Scenarios**:

1. **Dado** una pregunta por un artículo cuyo `graph check` sale con 0 y sin `version-obsoleta`, **Cuando** la skill responde, **Entonces** la respuesta lleva el texto citado, su cita y sus avisos de vigencia, y ninguna expresión de la lista (FR-040).
2. **Dado** una pregunta por un artículo cuyo `graph check` da un `version-obsoleta`, **Cuando** la skill responde, **Entonces** el cambio se dice solo con la línea `⚠ REDACCIÓN MODIFICADA:` y las dos fechas de vigencia tal como las da el hallazgo, y la respuesta no lleva la clase `version-obsoleta` ni la palabra «hallazgo» (FR-040, FR-042).
3. **Dado** un `graph check` que sale con un código distinto de 0, **Cuando** la skill responde, **Entonces** responde con el texto de `kitlegal boe` y dice que no se ha podido comprobar si la redacción ha cambiado desde una consulta anterior, sin nombrar la memoria de consultas, `graph` ni el código (FR-043).
4. **Dado** un bloque que `kitlegal boe` no puede leer porque la fuente no está disponible, **Cuando** la skill responde, **Entonces** dice qué no se pudo consultar y que la fuente no estaba disponible, sin el código de salida y sin suplir el texto (FR-044).

---

### User Story 2 - La eval de la consulta repetida plantea una situación que puede darse (Priority: P1)

La eval que mide la consulta repetida pregunta por el art. 118 LCSP, que el Real Decreto-ley 3/2020 modificó. Su grafo previo registra una lectura que vio la redacción original, derivada de la respuesta grabada del BOE quitándole solo la redacción posterior; la caché sirve la grabada, con la vigente. El grafo y el BOE cuentan la misma historia.

**Why this priority**: sin evidencia coherente, la eval mide cómo reacciona el modelo a una contradicción que en la realidad no se da.

**Independent Test**: `TestGrabacionesDerivadas` sobre la derivada nueva; la preparación de la sesión de la eval (`TestEvalsDelRepositorio`, subprueba `grafo-previo`); y la eval en el job.

**Acceptance Scenarios**:

1. **Dado** la sesión de la eval nueva —grafo con una lectura del bloque `a1-30` de `BOE-A-2017-12902` que vio la redacción de vigencia 20180309 y caché que sirve la grabada, de vigencia 20200206—, **Cuando** se hace su pregunta, **Entonces** la sesión pasa si ejecuta `kitlegal boe articulo BOE-A-2017-12902 a1-30`, ejecuta `kitlegal graph check` con la norma `BOE-A-2017-12902`, no ejecuta `kitlegal graph show`, y la respuesta lleva la cita `[BOE-A-2017-12902, bloque a1-30]`, la forma `⚠ REDACCIÓN MODIFICADA:` y ninguna expresión de la lista (FR-001 a FR-003).
2. **Dado** esa sesión preparada, **Cuando** se lee el bloque con la caché y se ejecuta `graph check` con la norma, **Entonces** sale con 0 y da un `version-obsoleta` sobre la redacción de 20180309, con las fechas 20180309 y 20200206 (FR-002).
3. **Dado** la derivada del grafo previo servida en lugar de la grabada, **Cuando** `boe` lee el bloque, **Entonces** da la redacción original que trae la grabada: su fecha de vigencia 20180309, su norma modificadora, su texto y su huella (FR-010, FR-011).
4. **Dado** una derivada de grafo previo con una fecha de vigencia, un texto o una huella que la grabada no trae, **Cuando** se ejecuta `TestGrabacionesDerivadas`, **Entonces** falla nombrando esa derivada (FR-012).
5. **Dado** el árbol tras el hito, **Cuando** se buscan la eval 19 anterior, su grafo previo, su derivada y su párrafo sintético, **Entonces** no existen, y ningún test ni código los nombra (FR-020).

---

### User Story 3 - Una expresión prohibida hace que la sesión no pase, sin juicio de ningún modelo (Priority: P1)

El informe del job busca en cada respuesta de las evals que activan `boe-legislacion` las expresiones de su lista, por la forma y sin modelo. Una respuesta con alguna no pasa, como una con una cita ausente; en las evals que deciden, eso decide con el umbral de siempre, y en las informativas se publica.

**Why this priority**: es lo que hace visible el defecto a un juez y lo que impide que vuelva.

**Independent Test**: tests de `Juzgar`, de formato y de calibrado; y el informe del job.

**Acceptance Scenarios**:

1. **Dado** una respuesta sin ninguna expresión de la lista, **Cuando** se juzga, **Entonces** no se encuentra ninguna y ese criterio no impide que la sesión pase (FR-052).
2. **Dado** una respuesta con «Sin hallazgos en la memoria de consultas», **Cuando** se juzga, **Entonces** se encuentra una expresión de la familia de la maquinaria y la sesión no pasa (FR-050, FR-052).
3. **Dado** una respuesta con «te confirmé», **Cuando** se juzga, **Entonces** se encuentra una expresión de la familia de lo dicho en otra conversación y la sesión no pasa (FR-050, FR-052).
4. **Dado** una respuesta que lleva una expresión de la lista con otras mayúsculas, con espacios de más o dentro de un énfasis de Markdown, **Cuando** se juzga, **Entonces** se encuentra igual (FR-051).
5. **Dado** una respuesta con la línea `⚠ REDACCIÓN MODIFICADA:` y sus dos fechas, y nada más de la lista, **Cuando** se juzga, **Entonces** no se encuentra ninguna expresión (FR-051).
6. **Dado** una eval de no activación de `boe-legislacion`, o cualquier eval de `legal-core`, **Cuando** se juzga, **Entonces** la lista no se aplica y la sesión se juzga como antes del hito (FR-052).
7. **Dado** una serie de una eval que decide con el modelo que decide, **Cuando** dos de sus tres sesiones no llevan ninguna expresión y cumplen lo demás, **Entonces** la serie pasa; **Cuando** dos de ellas llevan alguna, **Entonces** la serie no pasa y el veredicto es `fallo` (FR-054).
8. **Dado** cualquier job, **Cuando** se lee su informe, **Entonces** cada sesión publica las expresiones encontradas (vacía si ninguna) y cada modelo, cuántas de sus respuestas en evals que activan la skill llevan alguna y sobre cuántas (FR-053).
9. **Dado** una lista mal formada, **Cuando** se ejecuta `make ci` o el job, **Entonces** es un fichero mal formado: `make ci` falla nombrándolo y el informe lo lista entre los mal formados, con el veredicto `fallo` (FR-055).

---

### User Story 4 - Nada de lo dicho en otra conversación (Priority: P2)

La skill no sabe qué se respondió en otra conversación: `graph check` solo dice qué redacción se leyó antes. La respuesta no afirma, confirma ni desmiente lo dicho entonces, y sin `version-obsoleta` no dice nada de lo consultado antes, ni que cambió ni que no cambió.

**Why this priority**: es el fallo de la sesión que no pasa en H7.1; una sola respuesta de 93, pero afirma algo que la skill no puede saber.

**Independent Test**: la familia de lo dicho en otra conversación en la lista (FR-050) y el calibrado (SC-003); lo dicho con otras palabras queda visible en el informe, que publica cada respuesta (limitación declarada).

**Acceptance Scenarios**:

1. **Dado** una pregunta que dice que ya se preguntó antes («Hace tiempo te pregunté…») y un `graph check` con 0 y sin hallazgos —el bloque se leyó antes sin cambios o nunca se leyó—, **Cuando** la skill responde, **Entonces** da el texto citado y no afirma que la redacción cambió ni que no cambió desde entonces, ni lo que se respondió (FR-041).
2. **Dado** las 93 respuestas del informe de H7.1, **Cuando** se aplica la lista, **Entonces** la de «te habría confirmado antes» (eval 19, tercera sesión de Sonnet 5) queda marcada por la familia de lo dicho en otra conversación (FR-084).

---

### User Story 5 - Cualquiera repite la aceptación de la consulta repetida (Priority: P2)

Una persona, al leer el informe final, prepara un directorio de caché como la sesión de la eval nueva y hace la misma pregunta en dos conversaciones de Claude Code, una tras otra, con las órdenes del quickstart; una comprobación mecánica le dice si las dos respuestas son las esperadas.

**Why this priority**: la aceptación de H7.1 en Claude Code se describía, no se podía repetir tal cual.

**Independent Test**: el escenario del quickstart, fuera de `make ci` (FR-062).

**Acceptance Scenarios**:

1. **Dado** `KITLEGAL_CACHE_DIR` en un directorio temporal preparado como la sesión de la eval nueva, con el binario y la skill del hito instalados solo en ese directorio, **Cuando** se hace la pregunta de la eval en una conversación de Claude Code (una invocación no interactiva, `claude -p`, cuya salida queda en un fichero de ese directorio), **Entonces** la respuesta lleva `⚠ REDACCIÓN MODIFICADA:` en una línea con 20180309 y 20200206, y ninguna expresión de la lista (FR-060, FR-061).
2. **Dado** el estado que deja la primera conversación, **Cuando** se hace la misma pregunta en una segunda conversación nueva (otra invocación de `claude -p`, sin `--continue` ni `--resume`), **Entonces** la respuesta no lleva `⚠ REDACCIÓN MODIFICADA:` ni ninguna expresión de la lista (FR-061).
3. **Dado** las dos respuestas guardadas por las invocaciones, **Cuando** se ejecuta la comprobación del quickstart, **Entonces** dice si se cumplen los escenarios 1 y 2, y cuál no si alguno falla (FR-061); la persona no escribe ni copia ninguna respuesta a mano.

---

### Edge Cases

- **La pregunta dice «te pregunté» y el bloque nunca se leyó en este `KITLEGAL_CACHE_DIR`**: `graph check` no da nada; la respuesta no dice nada de la consulta anterior (FR-041).
- **La respuesta cita bloques de dos normas**: una comprobación por norma, como en H7.1 FR 040; la forma va una vez por cada bloque con `version-obsoleta` (FR-042, FR-047).
- **La sesión de la prueba de red** (la que el job añade con la etiqueta `evals-prueba-de-red` o la entrada `prueba_de_red`; su pregunta pide decir qué devolvieron dos órdenes que fallan): la lista se le aplica y sus expresiones se publican, pero no entra en el recuento por modelo (FR-053); si lleva alguna, esa sesión no pasa y no decide (FR-054).
- **La respuesta transcribe un artículo**: ninguna expresión de la lista está en el texto de los bloques grabados para las evals (FR-085); fuera de las evals la lista no se aplica (Fuera de alcance).
- **Las formas fijas de los avisos de vigencia** (`⚠ NORMA DEROGADA:`…) y frases como «No hay avisos de vigencia sobre este bloque» dicen derecho, no maquinaria: no son expresiones de la lista, y el calibrado (FR-084) lo comprueba sobre las dos respuestas de H7.1 que las llevan.
- **Lo mismo dicho con otras palabras** que la lista no recoge: no se detecta (limitación declarada, FR-051); el informe publica cada respuesta.
- **`graph check` o `kitlegal boe` fallan en una pregunta**: se dice lo que significa, sin el código (FR-043, FR-044); en la pregunta siguiente, si ya funcionan, no queda rastro.
- **Una lista mal escrita**: es un fichero mal formado (FR-055), sin casos propios.

## Requirements *(mandatory)*

### Functional Requirements

#### La eval de la consulta repetida

- **FR-001**: `evals/boe-legislacion/` MUST sustituir la eval 19 actual por una eval nueva, que ocupa el número 19, sobre el art. 118 de la LCSP (`BOE-A-2017-12902`, bloque `a1-30`, «Expediente de contratación en contratos menores»), con la pregunta literal «Hace tiempo te pregunté qué exige el artículo 118 de la LCSP para el expediente de un contrato menor. ¿Qué dice ahora?». Comprobable: el directorio tiene 19 ficheros de eval y el 19 lleva esa pregunta y activa la skill.
- **FR-002**: El grafo previo de la eval nueva MUST registrar una lectura del bloque `a1-30` de `BOE-A-2017-12902` que vio su redacción original (vigencia 20180309), preparada con la derivada de FR-010; la caché de la sesión MUST servir la respuesta grabada en H4, con la redacción vigente (vigencia 20200206, norma modificadora `BOE-A-2020-1651`); nada toca la red. Comprobable: en la sesión preparada, `kitlegal boe articulo BOE-A-2017-12902 a1-30` y después `kitlegal graph check BOE-A-2017-12902 a1-30` salen con 0, y `graph check` da exactamente un `version-obsoleta`, sobre la redacción de 20180309, con las fechas 20180309 y 20200206.
- **FR-003**: La eval nueva MUST exigir lo mismo que exigía la 19 (H7.1 FR 050): que la sesión ejecute `kitlegal boe articulo` de `BOE-A-2017-12902` `a1-30`, que ejecute `kitlegal graph check` con la norma `BOE-A-2017-12902`, que no ejecute `kitlegal graph show`, que la respuesta cite `[BOE-A-2017-12902, bloque a1-30]` y que lleve la forma de `version-obsoleta`, `⚠ REDACCIÓN MODIFICADA:`; y, como toda eval de la skill que la activa, que la respuesta no lleve ninguna expresión de la lista (FR-052).
- **FR-004**: La eval nueva MUST ser informativa (ADR 0016), como la 19. El informe del job MUST declarar para ella la forma `⚠ REDACCIÓN MODIFICADA:` y su tasa.

#### La derivada del grafo previo y el control de derivaciones

- **FR-010**: La derivada con la que se prepara el grafo previo de la eval nueva MUST ser la respuesta grabada en H4 del bloque `a1-30` de `BOE-A-2017-12902` sin su redacción posterior (la de vigencia 20200206) y sin ningún otro cambio: ninguna fecha, texto ni huella escritos a mano ni párrafo sintético. La produce código del repositorio a partir de la grabada, de forma reproducible, y entra, con la eval, por tareas `[datos]`.
- **FR-011**: `TestGrabacionesDerivadas` MUST exigir a toda derivada del grafo previo de una eval que, servida en lugar de la grabación de la que sale, `boe` dé exactamente una de las redacciones que trae la grabada: su fecha de vigencia, su norma modificadora, su texto y su huella. Comprobable: con la derivada de FR-010, `boe` da la fecha de vigencia 20180309 y la norma modificadora, el texto y la huella de la redacción original tal como los trae la grabada.
- **FR-012**: Una derivada del grafo previo de una eval con una fecha de vigencia, un texto o una huella que la grabada no trae en ninguna de sus redacciones MUST hacer fallar `TestGrabacionesDerivadas`, nombrándola. Comprobable: una derivada construida como la de la eval 19 retirada (fecha 20151002 y párrafo sintético) no pasa.
- **FR-013**: Las derivadas del e2e de H7 y H7.1 (`version-posterior`, `version-ulterior`, `sin-eli`, `eli-sin-segmento`) MUST seguir con su comprobación actual, sin cambios; y todo fichero de las carpetas de derivadas MUST seguir teniendo su comprobación, y toda comprobación su fichero.
- **FR-014**: Si preparar la eval nueva necesita alguna consulta que aún no esté grabada, MUST grabarla el paso `grabar_datos` del workflow, de la fuente `boe.legislacion-consolidada` (fila revisada en `docs/SOURCES.md`); ninguna tarea toca la red.

#### Lo que se retira

- **FR-020**: MUST retirarse `evals/boe-legislacion/19-lpac-articulo-21-redaccion-cambiada.yaml`, `testdata/evals/grafo-previo/lpac-a21-version-anterior/`, su entrada en `grabacionesDerivadas()` y su párrafo sintético (`parrafoDeLaVersionAnterior`), en `internal/app/grafo_test.go`. Comprobable: no existen, y ningún fichero de código ni de test del repositorio los nombra.
- **FR-021**: Los tests que nombran ese fichero o esa derivada MUST adaptarse para seguir comprobando lo mismo con la eval o la derivada nuevas; ninguno se desactiva, se salta ni se retira salvo lo que retira FR-020.

#### Qué evals deciden

- **FR-030**: El conjunto de `boe-legislacion` MUST quedarse en 19 evals, con exactamente 10 positivas que deciden (de la 01 a la 10); siguen informativas de la 13 a la 18, y la nueva 19 lo es (FR-004); y las reglas del conjunto no cambian. Comprobable: `make ci` con las reglas del conjunto.

#### Skill `boe-legislacion` v0.1.2

- **FR-040**: `skills/boe-legislacion/SKILL.md` MUST pedir que la respuesta, salvo la forma `⚠ REDACCIÓN MODIFICADA:` cuando `graph check` da `version-obsoleta`, no mencione nada de esto: la memoria de consultas, `graph check` ni ningún verbo de `kitlegal graph`, los códigos de salida, los hallazgos, las clases del binario (`version-obsoleta`, `fuente-caducada`), el JSON ni el sobre.
- **FR-041**: `SKILL.md` MUST pedir que la respuesta no afirme, confirme ni desmienta lo que se dijo o se respondió en otra conversación, y que sin `version-obsoleta` no afirme nada sobre lo que se consultó antes, ni que cambió ni que no cambió, porque `graph check` sin hallazgos no distingue un bloque leído antes y sin cambios de uno que nunca se leyó.
- **FR-042**: Con `version-obsoleta`, `SKILL.md` MUST mantener la forma fija de H7.1 FR 042 y pedir que las dos fechas de vigencia de la línea —la de la redacción superada y la de la leída— vayan tal como las da el hallazgo (`AAAAMMDD`), en la misma línea que la forma. Comprobable: en el escenario del quickstart, la línea lleva 20180309 y 20200206 (FR-061).
- **FR-043**: La regla 7 de `SKILL.md` MUST reescribirse para que, cuando `graph check` sale con un código distinto de 0, la skill responda igualmente con el texto de `kitlegal boe` y diga que no se ha podido comprobar si la redacción ha cambiado desde una consulta anterior, sin afirmar que cambió ni que no cambió (FR-041) y sin nombrar la memoria de consultas, `graph` ni el código. Ni la regla 7 ni el ejemplo de la línea `⚠ REDACCIÓN MODIFICADA:` MUST pedir a la respuesta ninguna expresión de la lista.
- **FR-044**: `SKILL.md` MUST pedir que lo que no se pudo consultar con `kitlegal boe` se diga por lo que significa para quien pregunta —que la fuente no estaba disponible, que la fuente limitó las consultas, que el artículo no está en la norma—, sin el código de salida; lo demás de la regla 2 (no suplir el texto, pedir los bloques por separado, decir cuáles faltan) se queda.
- **FR-045**: El cambio de `SKILL.md` que corrige el ruido MUST trazarse a su causa de raíz, que el research explica con la evidencia de los informes de H7 y H7.1: el ruido es solo de Sonnet 5, va siempre en el primer párrafo y sigue a la última orden, que desde H7.1 es `graph check`, pese a la regla vigente. Una prohibición más, sin explicar por qué la actual no basta, no satisface este requisito. Comprobable: el research da la causa con esas medidas y el plan traza a ella cada cambio de `SKILL.md`; su efecto lo mide SC-001.
- **FR-046**: Se quedan en `SKILL.md`: H7 FR 081 (el texto citado sale de `kitlegal boe`, nunca de `graph`); la parte de H7 FR 082 que trata el código 0 como un resultado; H7.1 FR 040, FR 041 y FR 043 (una comprobación por norma citada, después de leer; cada bloque una vez por pregunta; `fuente-caducada` no se traslada); H7.1 FR 045 (la etiqueta la da el binario, con su comprobación mecánica en `make ci`); y la forma de la cita y la de los avisos de vigencia.
- **FR-047**: `SKILL.md` MUST NOT cambiar más allá de lo que piden FR-040 a FR-045; MUST seguir por debajo de 300 líneas, con frontmatter válido, la región generada intacta (`make skills-check` sin drift) y sin nombrar evals, el job ni modelos (H5 FR 077).
- **FR-048**: `CHANGELOG.md` (*Unreleased*) MUST registrar `boe-legislacion` v0.1.2 con lo que cambia para quien la usa.

#### La lista de expresiones prohibidas

- **FR-050**: El formato común de eval MUST poder expresar, una por skill, una lista de expresiones prohibidas en la respuesta, y `boe-legislacion` MUST tener la suya, con dos familias:
  - **la maquinaria interna**: la memoria de consultas, `graph check` y los demás verbos de `kitlegal graph`, los códigos de salida, los hallazgos, las clases `version-obsoleta` y `fuente-caducada`, el JSON y el sobre, en formas que no chocan con el castellano corriente de una respuesta ni con el texto de las normas (FR-084, FR-085); un elemento que no tenga una forma así queda, como las paráfrasis, fuera de la comprobación y visible en el informe (FR-051);
  - **lo dicho en otra conversación**: las expresiones que atribuyen a la skill algo dicho a quien pregunta en otro momento, un verbo de decir en pasado o en condicional con «te» («te dije», «te confirmé», «te habría confirmado»…) o «conversación anterior».

  El plan fija las expresiones exactas y dónde vive la lista. Ninguna otra skill tiene lista, y las evals de una skill sin lista se leen y se juzgan como antes del hito. La documentación del formato común de eval (`CONTRIBUTING.md`, «Formato común de eval») la describe.
- **FR-051**: Cada expresión MUST compararse con la respuesta por la forma, sin juicio de ningún modelo, con la tolerancia de H5.1 a mayúsculas, espacios y énfasis de Markdown. La forma `⚠ REDACCIÓN MODIFICADA:` y las formas de los avisos de vigencia MUST NOT ser expresiones prohibidas. Lo dicho con otras palabras que la lista no recoge no se detecta: es una limitación declarada, visible en el informe, que publica cada respuesta.
- **FR-052**: `Juzgar` MUST aplicar la lista de la skill a cada sesión de cada eval de esa skill que activa la skill, con los dos modelos: una respuesta que lleva una o más expresiones de la lista no pasa, como una con una cita ausente, y una que no lleva ninguna no deja de pasar por esto. No se aplica a las evals de no activación.
- **FR-053**: El informe del job MUST publicar, por sesión, las expresiones de la lista encontradas en su respuesta (vacío si ninguna) y, por modelo, cuántas de sus respuestas en evals que activan la skill llevan alguna y sobre cuántas respuestas. El recuento por modelo se hace sobre las sesiones de las series que pide el plan de sesiones; la sesión de la prueba de red, serie no planificada que no decide, se juzga con la lista y publica sus expresiones encontradas como cualquier sesión, pero MUST NOT entrar en el recuento por modelo (con el conjunto actual, 51 respuestas de Sonnet 5 y 30 de Haiku 4.5, con o sin ella). Comprobable: un job con la prueba de red activa da el mismo recuento por modelo que sin ella.
- **FR-054**: La comprobación MUST decidir en las 10 positivas que deciden con el umbral del ADR 0016 (2 de 3 sesiones por serie): una sesión con una expresión es una sesión que no pasa. En las informativas, lo mismo cuenta en su tasa, que se publica sin decidir. La regla del veredicto del ADR 0016 no cambia en nada más.
- **FR-055**: Una lista que no cumple su formato MUST ser un fichero mal formado, como un fichero de eval: `make ci` falla nombrándolo, y el job lo lista entre los mal formados y da el veredicto `fallo`, el que ya da (ADR 0016).
- **FR-056**: `internal/evals` y `schemas/eval.yaml.json` MUST NOT cambiar más allá de la lista (FR-050 a FR-055), de la derivación real del grafo previo (FR-010 a FR-012) y del punto de entrada de prueba de la comprobación del quickstart (FR-061), que aplica la lista y su comparación a dos respuestas y no añade criterios a `Juzgar`, formato de eval ni campos al informe. `schemas/eval.yaml.json` y lo que viva bajo `testdata/` entran por tareas `[datos]`.

#### Escenario de quickstart

- **FR-060**: `quickstart.md` MUST tener un escenario que, con el binario y la skill del hito y `KITLEGAL_CACHE_DIR` en un directorio temporal preparado como la sesión de la eval nueva (el grafo con la redacción original y la caché con la grabada, con la misma preparación que el job), haga la pregunta de la eval en dos conversaciones de Claude Code, una tras otra. Cada conversación MUST ser una invocación no interactiva (`claude -p`) con la pregunta literal de la eval nueva, sin `--continue` ni `--resume` (la segunda es una conversación nueva), lanzada desde el directorio de trabajo del escenario y con `KITLEGAL_CACHE_DIR` en el directorio temporal preparado, y MUST escribir su salida en un fichero propio de ese directorio, del que la comprobación de FR-061 lee la respuesta final de la sesión, la misma que juzgaría el job; la persona solo ejecuta las órdenes, sin escribir ni copiar nada a mano. El binario MUST construirse desde la rama del hito dentro del directorio temporal y ir primero en el `PATH` de las órdenes, y la skill MUST instalarse con ese binario, `kitlegal skills install --host claude`, en ámbito local, en el directorio de trabajo de las conversaciones. Sus órdenes MUST poder repetirse tal cual y MUST NOT cambiar, fuera del directorio temporal, el árbol de trabajo, el índice ni el historial de git, la instalación de kitlegal ni las skills de la persona, ni dejar conversaciones guardadas fuera de él. El plan MUST fijar las banderas exactas (modelo, permisos justos para que la sesión ejecute `kitlegal`, formato de salida y de dónde se toma la respuesta final) y cómo se garantiza que la conversación carga la skill instalada en ese directorio y no otra con el mismo nombre de la cuenta de la persona.
- **FR-061**: El escenario MUST fijar una comprobación mecánica de las dos respuestas: la primera lleva `⚠ REDACCIÓN MODIFICADA:` en una línea con 20180309 y 20200206; la segunda no la lleva; y ninguna de las dos lleva una expresión de la lista, comparada como en FR-051. La comprobación dice cuál de las tres condiciones falla, si alguna. MUST ser un punto de entrada de prueba de `internal/evals` con la etiqueta de compilación `evals`, invocado desde el quickstart con `go test -tags evals -run … ./internal/evals/ -args …` (como `TestPrepararSesion`, que el escenario ya usa para preparar el directorio), que lee las dos respuestas guardadas y usa el mismo código que el job para la lista y su comparación (FR-051) y para la forma `⚠ REDACCIÓN MODIFICADA:`; sin réplicas de la lista o de la comparación en `quickstart.md` ni en un guion aparte. El plan fija su nombre y sus banderas.
- **FR-062**: El escenario MUST quedar fuera de `make ci` y del run —necesita modelo— y lo ejecuta la persona al leer el informe final.

#### Relación con otros hitos

- **FR-070**: `specs/010-h7-internal-graph-grafo/` y `specs/011-h7-1-graph-check-acotado/` MUST NOT editarse; este spec nombra lo que sustituye de ellos («Relación con H7 y H7.1»). No hay ADR nuevo.

#### Controles y Definition of Done

- **FR-080**: `make ci` MUST quedar en verde, con `skills-check` y `schema-check` sin drift y con las reglas del conjunto.
- **FR-081**: `TestGrabacionesDerivadas` MUST cumplir FR-011 a FR-013.
- **FR-082**: Los tests de formato MUST cubrir que una lista mal formada es un fichero mal formado (FR-055).
- **FR-083**: Los tests de `Juzgar` MUST cubrir una respuesta sin expresiones prohibidas; una con una expresión de cada familia; las variantes toleradas (mayúsculas, espacios y énfasis de Markdown); y una con la forma `⚠ REDACCIÓN MODIFICADA:`, que no es una expresión prohibida.
- **FR-084**: Una comprobación de calibrado MUST aplicar la lista, con la comparación de FR-051, a las 93 respuestas del informe de H7.1 (`specs/011-h7-1-graph-check-acotado/gates/evals/boe-legislacion.json`) y exigir que marque exactamente las 35 que cuenta la entrada de `docs/USO.md` del 2026-09-29 —las 34 que marca su patrón de ruido interno y la que marca su patrón de lo dicho en otra conversación (la de «te habría confirmado antes», eval 19)—, eval por eval: 02 (1), 03 (3), 04 (3), 05 (3), 06 (3), 07 (3), 08 (2), 09 (2), 13 (3), 14 (3), 15 (3), 16 (2), 17 (3) y 19 (1); y ninguna otra.
- **FR-085**: Una comprobación MUST exigir que ninguna expresión de la lista aparezca, con la comparación de FR-051, en el texto de ningún bloque grabado para las evals de `boe-legislacion`, incluidas las dos redacciones del bloque `a1-30` de la LCSP.
- **FR-086**: `CHANGELOG.md` (*Unreleased*) MUST registrar la eval nueva de la consulta repetida, la lista de expresiones prohibidas y su decisión, lo retirado y `boe-legislacion` v0.1.2 (FR-048).
- **FR-087**: El job de evals MUST ejecutarse en la propuesta de cambio (lo hace el workflow tras la revisión final), y su informe da lo que pide SC-001.

### Key Entities

- **Lista de expresiones prohibidas**: una por skill, en el formato común de eval; cada expresión es de una de dos familias (la maquinaria interna, lo dicho en otra conversación).
- **Expresión encontrada**: una expresión de la lista presente en una respuesta, con la tolerancia de H5.1; se publica por sesión.
- **Recuento por modelo**: cuántas respuestas de un modelo en evals que activan la skill llevan alguna expresión, sobre cuántas.
- **Derivada del grafo previo**: una respuesta grabada del BOE sin alguna de sus redacciones; servida en lugar de la grabada, da una redacción que la grabada trae.
- **Redacción de la grabada**: cada `<version>` de la respuesta de un bloque, con su fecha de vigencia, su norma modificadora, su texto y su huella.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: En el informe del job de cierre, la eval nueva declara `⚠ REDACCIÓN MODIFICADA:` y su tasa, cada sesión publica sus expresiones prohibidas, y el recuento de cada modelo queda dentro del umbral: como mucho el 5 % de sus respuestas en las evals que activan la skill llevan alguna —con el conjunto actual, ≤ 2 de 51 con Sonnet 5 y ≤ 1 de 30 con Haiku 4.5, sin contar la sesión de la prueba de red (FR-053)—, frente a 34 de 51 (67 %) con Sonnet 5 en H7.1; con el veredicto aprobado (el hito cambia `SKILL.md`, Definition of Done §1.10) y `red` vacío.
- **SC-002**: El escenario del quickstart da `⚠ REDACCIÓN MODIFICADA:` en una línea con 20180309 y 20200206 en la primera conversación y no en la segunda, con 0 expresiones prohibidas en las dos, y su comprobación mecánica lo confirma.
- **SC-003**: Sobre las 93 respuestas del informe de H7.1, la lista marca exactamente 35, con el reparto por eval de FR-084, y 0 de las otras 58.
- **SC-004**: 0 expresiones de la lista en el texto de los bloques grabados para las evals de `boe-legislacion`.
- **SC-005**: `TestGrabacionesDerivadas` pasa con la derivada nueva, que da la redacción de vigencia 20180309 con la norma modificadora, el texto y la huella que trae la grabada, y falla con una derivada de grafo previo con una fecha, un texto o una huella inventados.
- **SC-006**: Los tests de `Juzgar` distinguen los cuatro casos de FR-083, y los de formato dan por mal formada una lista mal formada.
- **SC-007**: En el árbol tras el hito, 0 ficheros de la eval 19 anterior y de su grafo previo, 0 menciones de ellos o de su derivada en código y tests, 19 evals en `evals/boe-legislacion/` con 10 positivas que deciden.
- **SC-008**: `make ci` en verde, con `skills-check` y `schema-check` sin drift y con las reglas del conjunto; `SKILL.md` de `boe-legislacion` por debajo de 300 líneas; y la entrada de `CHANGELOG.md` (*Unreleased*).

## Uso, de fuera adentro

Cada salida del hito, desde quien la consume (criterio de uso, ADR 0028). El hito no cambia el binario: lo que la skill lee de `graph check` sigue acotado como en H7.1 (una comprobación por norma citada, ≤ k hallazgos con k bloques leídos, ≤ 3 800 bytes con cinco bloques cambiados y ≈ 300 sin cambios, con cualquier volumen; H7.1 SC 005). Volumen de referencia: meses de uso diario, cientos de normas y miles de bloques consultados, la mayoría hace más de una semana.

| Salida | Quién la pide, cuántas veces y qué hace con ella | Tamaño con meses de uso | Cuándo deja de darse cada señal |
|---|---|---|---|
| La respuesta de `boe-legislacion` | La persona que pregunta; una por pregunta. Lee el texto citado, sus avisos y, si cambió, la línea `⚠ REDACCIÓN MODIFICADA:` (FR-040 a FR-044). | Lo que ocupa la norma: 0 líneas sobre la comprobación (FR-040); una línea de forma de ≈ 150 bytes por bloque leído cuya redacción cambió —0 en la mayoría de las preguntas, como mucho k × 150 bytes con k bloques leídos—, y a lo sumo una frase de ≈ 100 bytes por fallo de `graph check` o `kitlegal boe` (FR-043, FR-044). No crece con lo acumulado: con cientos de normas y miles de bloques consultados, la respuesta es la misma, porque no cuenta la memoria. | La línea `⚠ REDACCIÓN MODIFICADA:` sale en la respuesta de la lectura que ve la redacción nueva y no en la siguiente sobre ese artículo, que la apaga (H7.1 FR 024; SC-002); al cabo de un mes, solo si el BOE publica otra redacción. La frase de un fallo solo sale en la respuesta en que falló. Nada sobre consultas anteriores sin `version-obsoleta` (FR-041). |
| Las expresiones encontradas de cada sesión (informe) | El job, que juzga cada sesión con ellas (FR-052, FR-054), y la persona que lee el informe final; una vez por sesión y job. | Por sesión, como mucho las L expresiones de la lista (la fija el plan, del orden de decenas) × ≈ 30 bytes, ≤ 1 KB; lo habitual, 0. Por job, 93 sesiones × ≤ 1 KB ≤ 93 KB en el peor caso. No depende del uso del kit: la lista es fija. | Cada job las mide de nuevo sobre su commit; no se acumulan entre jobs. |
| El recuento por modelo (informe) | La persona que lee el informe final y comprueba el umbral de SC-001; una vez por job. | Dos enteros por modelo (respuestas con alguna expresión, respuestas en evals que activan). | Como arriba. |
| La lista de expresiones prohibidas | El job, al juzgar cada sesión de las evals que activan la skill; la comprobación de calibrado (FR-084) y la de los bloques grabados (FR-085) en `make ci`. | Fija, del orden de decenas de expresiones; no crece con el uso. | No da señales. |
| La eval nueva y su grafo previo | El job; tres sesiones por job con el modelo que decide (informativa, FR-004). | Un bloque en el grafo previo y una grabada en la caché. | Su `version-obsoleta` se da en cada sesión porque cada una empieza del mismo estado preparado; la señal se apaga dentro de la sesión con una segunda lectura, que el protocolo no hace (H7.1 FR 041). |
| El escenario del quickstart | La persona, al leer el informe final; una vez, dos conversaciones. | Dos respuestas y una comprobación. | La forma sale en la primera conversación y no en la segunda (FR-061). |

## Fuera de alcance

Del hito (literal arriba): cambiar el binario (`graph check`, sus clases, su salida) o el protocolo de la skill más allá de lo que piden el Alcance y la regla 7; tocar `internal/evals` y `schemas/eval.yaml.json` más allá de la lista de expresiones prohibidas y de la derivación real del grafo previo; promover a decisoria ninguna eval informativa; extender la lista a `legal-core` o a otras skills; juzgar la redacción libre de la respuesta con un modelo o por similitud, incluidas las paráfrasis que la lista no recoge; la misma pregunta dos veces dentro de una misma conversación; y editar el spec, el plan, los guiones o las grabaciones de H7 y H7.1 fuera de lo que retira el Alcance.

De lo que el hito no especifica (constitución, «Criterio de decisión autónoma», punto 2):

- Que el umbral del 5 % por modelo entre en el veredicto del job: el veredicto sigue la regla del ADR 0016 (FR-054) y el umbral se comprueba en el recuento que publica el informe (SC-001).
- Aplicar la lista fuera de las evals —filtrar o comprobar las respuestas en el uso real, en el binario o en la skill—.
- Cambiar la eval 02 o cualquier otra eval distinta de la 19, salvo lo que les añade la lista al juzgarlas.
- Cambiar el informe del job más allá de las expresiones por sesión y el recuento por modelo (FR-053).
- Contar la sesión de la prueba de red en el recuento por modelo, o incluirla en la ejecución de cierre (`scripts/workflow/cierre.sh` pone solo la etiqueta `evals`).
- Una forma interactiva del escenario del quickstart (sesión de `claude` en la que la persona copia la respuesta a un fichero) o dos formas aceptadas por la comprobación.
- Instalar el binario o la skill en ámbito global (`-g`, `make install`) para el escenario del quickstart.
- Comprobar las dos respuestas del quickstart con órdenes de shell replicadas en `quickstart.md` o con un guion nuevo bajo `scripts/`.
- Ejecutar el escenario del quickstart en `make ci`, en el run o en el job.
- Cambiar las formas de la cita y de los avisos de vigencia, o escribir en la línea `⚠ REDACCIÓN MODIFICADA:` las fechas en otro formato que el del hallazgo.
- Cualquier grabación nueva distinta de la que la preparación de la eval nueva necesite (FR-014).

## Assumptions

- La evidencia de H7.1 es su informe versionado (`specs/011-h7-1-graph-check-acotado/gates/evals/boe-legislacion.json`, 93 respuestas); la de H7 es la medida de `docs/USO.md` (entrada del 2026-09-29), sacada del registro del job 36431889063, que no está versionado.
- La respuesta grabada en H4 del bloque `a1-30` (`internal/source/boe/testdata/boe.legislacion-consolidada/`) trae dos redacciones: la original, de `BOE-A-2017-12902`, publicada el 20171109 con vigencia 20180309, y la de `BOE-A-2020-1651`, publicada el 20200205 con vigencia 20200206; `boe articulo` da la última. Quitar la posterior deja la original como la que da `boe`.
- Con el modelo que decide se ejecutan las 19 evals y con Haiku 4.5 solo las que no son informativas (ADR 0016): 51 y 30 respuestas en evals que activan la skill, con 3 repeticiones.
- La pregunta de la eval nueva es el ejemplo literal del hito.
- La preparación del directorio del quickstart es la misma que la de una sesión del job; la persona tiene Claude Code y su credencial; el binario y la skill los instala el propio escenario, solo en el directorio temporal (FR-060).
- Los nombres técnicos que aparecen (`Juzgar`, `TestGrabacionesDerivadas`, `grabacionesDerivadas()`, `parrafoDeLaVersionAnterior`, los ficheros que se retiran, `schemas/eval.yaml.json`) los fija el hito o existen en el repositorio; las expresiones exactas de la lista, dónde vive, su formato, los nombres de campos del informe y el nombre del fichero de la eval nueva son del plan.
