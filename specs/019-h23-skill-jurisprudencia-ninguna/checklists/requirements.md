# Specification Quality Checklist: H23 · skill `jurisprudencia`: ninguna sentencia citada sin el documento que trae la persona + `cita preparar` y `cita cotejar`

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-10-07
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec dice qué recibe quien prepara una consulta o coteja un documento, qué dice la skill y qué mide el job. Los nombres de paquetes, ficheros, umbrales y tests que aparecen son los que fija el hito o existen en el repositorio; las claves de `data`, los argumentos de las herramientas, el identificador del hallazgo, los valores de `fuente` y `url` y los campos nuevos del formato de eval quedan para el plan (Assumptions, última viñeta).
- [x] Focused on user value and business needs — Resumen y US1 a US3: quien recibe una respuesta con una sentencia citada puede fiarse de que no es inventada, y quien no tiene el documento recibe la consulta exacta para traerlo.
- [x] Written for non-technical stakeholders — cada historia empieza por lo que ve una persona; el detalle técnico está en los requisitos, como en los specs de H21 y H24.
- [x] All mandatory sections completed — User Scenarios, Requirements, Success Criteria, Assumptions y «Fuera de alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — las tres marcas (FR-012, FR-043 y FR-048) se resolvieron en `## Clarifications` (sesión 2026-10-07: cuatro respuestas decididas, dos conservadoras) y el spec las integra en FR-012, FR-023, FR-043, FR-048, FR-050 (f), FR-065, «Fuera de alcance» y «Assumptions».
- [x] Requirements are testable and unambiguous — cada FR lleva su «Comprobable», su **Control**, un escenario de US1 a US6 o un control de FR-080 a FR-088. Las formas de las referencias están en FR-005 con ejemplos; lo que hace reconocible una ficha, en FR-021; cuándo un documento es el pedido y qué dice el hallazgo, en FR-024 y FR-025; y lo que cuenta en el umbral, en FR-061, con la regla que dice dónde acaba un ECLI dentro de un texto —la misma en la respuesta, en la pregunta y en la salida de una operación— y su caso en FR-086. Las dos reglas generales que chocaban con el Tribunal Constitucional nombran su excepción: FR-010 remite a FR-014, y FR-042, a FR-047, que dice que esa sentencia no lleva la línea ni pasa por `cita preparar`. Los tres puntos que no lo eran ya están resueltos (ítem anterior).
- [x] Success criteria are measurable — SC-001 y SC-003 a SC-009, con recuentos y códigos; SC-002 es la parte humana de la Aceptación y SC-010, la documentación.
- [x] Success criteria are technology-agnostic (no implementation details) — miden respuestas, códigos, cambios en el equipo y umbrales; los controles nombran la comprobación que falla, como exige el ADR 0029.
- [x] All acceptance scenarios are defined — US1 a US6, en forma Dado/Cuando/Entonces (29 escenarios: 6, 6, 7, 4, 5 y 1).
- [x] Edge cases are identified — «Edge Cases»: ninguna o varias referencias, `--fecha` sin `--resolucion`, el Constitucional nombrado sin su ECLI, texto delante o detrás de la ficha, una ficha incompleta, el fallo sin encabezamiento, el PDF, las sentencias que nombra el documento, el cruce entre ROJ y número de resolución, el agente sin herramienta ni binario y las llamadas simultáneas.
- [x] Scope is clearly bounded — «Fuera de alcance», con las doce viñetas literales del hito y lo no especificado.
- [x] Dependencies and assumptions identified — «Assumptions» y «Relación con otros hitos».

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — tabla de trazabilidad de «Criterios del hito, literales» y controles FR-080 a FR-088. FR-045 y FR-065 declaran que «resume» no tiene control en este hito, como pide el hito.
- [x] User scenarios cover primary flows — preparar la consulta (US1), cotejar el documento (US2), la respuesta de la skill (US3), el applet sin red y como herramienta (US4), el job y su umbral (US5) y la documentación (US6).
- [x] Feature meets measurable outcomes defined in Success Criteria — SC-001 y SC-002 recogen la Aceptación literal; SC-003 a SC-009, los Controles.
- [x] No implementation details leak into specification — ver el primer ítem.

## Criterios del juez del spec (workflow `hito`, `juez_spec`)

- [x] a. Alcance: cubre la Entrega y el Alcance del hito y nada más (tabla de trazabilidad). Los estados que no pasan el umbral de materialidad —ficheros del kit manipulados a mano, un documento falso— van a la regla genérica y al ADR 0036 («Edge Cases», última viñeta).
- [x] b. Fuera de alcance: las doce viñetas literales del hito, generadas desde `docs/ROADMAP.md` con `awk` y comprobadas una a una contra él, y lo no especificado.
- [x] d. Aceptación: «Criterios del hito, literales» lleva Entrega, Controles y Aceptación transcritos desde `docs/ROADMAP.md` con `awk` (14 líneas, comprobadas una a una); SC-001 a SC-010.
- [x] e. Decisiones cerradas: kitlegal no consulta el CENDOJ de ninguna manera (FR-002); ninguna sentencia se cita sin su documento ni se dice que existe (FR-041, FR-042); solo GET y HEAD, y aquí ni eso (FR-002, FR-084); el sobre (FR-004); el contrato de resultados por filas, con el hallazgo en `data` y código 0 (FR-025) y los errores de argumentos con 2 (FR-006); skills sin código y con las dos formas en la tabla (FR-040); y el juez, nunca sin medir (FR-055, FR-065).
- [x] g. Sin implementación: las formas concretas son del plan; los nombres que aparecen los da el hito o existen en el repositorio.
- [x] i. Uso: «Uso, de fuera adentro», con quién pide cada salida y cuántas veces, su tamaño con su cálculo —ninguna crece con el uso, porque el applet no guarda nada (FR-003)— y cuándo deja de darse cada señal.
- [x] j. Umbral con control: `cita_sin_documento` y `sin_activar` nombran el job en rojo (FR-060, FR-062); las 300 líneas, `skills-check` (FR-040); las doce herramientas, el control de conformidad de H21 (FR-030); los cuatro umbrales del informe y las seis evals, sus tests de `make ci` (FR-063, FR-056); los 2 353 bytes del fragmento, su test (FR-053); SC-001 y SC-003 a SC-009 llevan **Control**; y FR-066 fija las filas de «Controles de umbral». «Resume» no es un requisito con umbral: el hito lo deja sin control y lo anota como supuesto (FR-065).

## Notes

- Los criterios c (clarificaciones), f (checklist veraz) y h (derivable) del juez no se marcan aquí: son del juez. Las tres marcas [NEEDS CLARIFICATION] que los frenaban se cerraron en el paso de clarify (sesión 2026-10-07); las lecturas que el spec fija salen del hito, del ADR 0036, de la constitución o de esas respuestas, y están en Assumptions con la alternativa rechazada. Dos de ellas son conservadoras: que el equivalente entre ECLI y ROJ solo se deduce para el Tribunal Supremo y que la skill no cita un documento cuyo ROJ y ECLI no se corresponden; la segunda no tiene control en este hito (FR-065).
- Corrección tras el primer veredicto del juez (`gates/spec-r1.json`, dos motivos): la delimitación de un ECLI (FR-061, FR-086) y la excepción del Tribunal Constitucional (FR-010, FR-011, FR-042, FR-043, FR-047, la fila (e) de FR-050, el escenario 6 de US3, «Edge Cases», la fila de FR-014 de «Uso» y Assumptions); de la misma clase, FR-013 remite a FR-015 para el texto vacío. Lo que se decidió está en `gates/supuestos.md`. En la sesión del corrector solo se han ejecutado lecturas y recuentos sobre el spec y `docs/ROADMAP.md`: ningún test ni `make ci`.
- En la sesión que escribió el spec no se ha ejecutado ningún test ni `make ci`, ni se ha consultado ninguna fuente. Lo ejecutado: lecturas y búsquedas en el repositorio y en la rama de lectura; la transcripción literal con `awk` y su comprobación; el recuento de ids —65 definidos, ninguno duplicado y ninguno citado sin definir—, de escenarios y de marcas; el cálculo de tamaños con `jq` y `wc -c`; y `kitlegal territorio resolver Leganés --json` con el binario que había compilado en `bin/` (`v0.3.2-22-g4152c2b`, anterior a la cabeza de `main`), para medir lo que el sobre añade a `data`.
