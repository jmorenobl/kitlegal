# Specification Quality Checklist: H23 · `cita resolver`: comprobar que una sentencia existe, por el formulario del CENDOJ + skill `jurisprudencia`

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-10-06
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec dice qué recibe quien consulta, qué dice la skill y qué mide el job. Los nombres de paquetes, ficheros, campos del formulario, umbrales y tests que aparecen son los que fija el hito o existen en el repositorio; las claves de `data`, los argumentos de la herramienta, cómo declara una fuente su formulario y los campos nuevos del formato de eval quedan para el plan (Assumptions, última viñeta).
- [x] Focused on user value and business needs — Resumen y US1 a US3: quien recibe una respuesta con una sentencia citada puede fiarse de que existe.
- [x] Written for non-technical stakeholders — cada historia empieza por lo que ve una persona; el detalle técnico está en los requisitos, como en H21 y H24.
- [x] All mandatory sections completed — User Scenarios, Requirements, Success Criteria, Assumptions y «Fuera de alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — las tres marcas del spec inicial se resolvieron en `## Clarifications` (sesión 2026-10-06): FR-024 (el ritmo rige dentro de cada invocación), FR-041 (la caché guarda el «no encontrado» 30 días) y FR-067 (con más de una resolución, la skill no cita ninguna y pide el ECLI o el ROJ).
- [x] Requirements are testable and unambiguous — cada FR lleva su «Comprobable», su **Control**, un escenario de US1 a US10 o un control de FR-110 a FR-118. Las formas de las referencias están en FR-003 a FR-005 con ejemplos que se aceptan y que no. Tras el rechazo del juez (ronda 1): el mensaje del «no encontrado» dice qué nombra y qué no lleva, sin contradecirse en `--roj` con `--fecha` (FR-012, US2 escenario 2), y toda respuesta de la página o del formulario tiene fila: estado 200 reconocido, 5xx (4) y cualquier otro estado o respuesta (5), con la redirección sin seguir (FR-016, FR-021, FR-022, FR-033).
- [x] Success criteria are measurable — SC-001 a SC-007 y SC-009, con recuentos y códigos; SC-008 es la parte humana de la Aceptación.
- [x] Success criteria are technology-agnostic (no implementation details) — miden respuestas, códigos, peticiones y umbrales; los controles nombran la comprobación que falla, como exige el ADR 0029.
- [x] All acceptance scenarios are defined — US1 a US10, en forma Dado/Cuando/Entonces (44 escenarios: 6, 5, 6, 5, 4, 6, 5, 3, 3 y 1).
- [x] Edge cases are identified — «Edge Cases»: ninguna o varias formas de referencia, `--fecha` con un ECLI, el Constitucional pedido sin su ECLI, el 403 antes del formulario, un resultado sin sus datos, el plazo, la caché vencida, el ROJ en el lugar del número, el texto pegado que no se resuelve y el agente sin herramienta ni binario.
- [x] Scope is clearly bounded — «Fuera de alcance», con las diez viñetas literales del hito y lo no especificado.
- [x] Dependencies and assumptions identified — «Assumptions» y «Relación con otros hitos».

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — tabla de trazabilidad de «Criterios del hito, literales» y controles FR-110 a FR-118; los tres FR que tenían marca, resueltos en clarify.
- [x] User scenarios cover primary flows — la herramienta (US1, US2), la skill (US3, US8), el formulario (US4), el bloqueo (US5), lo que se guarda (US6), el job (US7), la fuente a petición y la grabación (US9) y la documentación (US10).
- [x] Feature meets measurable outcomes defined in Success Criteria — SC-001 y SC-008 recogen la Aceptación literal; SC-002 a SC-007 y SC-009, los Controles.
- [x] No implementation details leak into specification — ver el primer ítem.

## Criterios del juez del spec (workflow `hito`, `juez_spec`)

- [x] a. Alcance: cubre la Entrega y el Alcance del hito y nada más (tabla de trazabilidad). Los estados que no pasan el umbral de materialidad —ficheros del kit manipulados a mano— van a la regla genérica («Edge Cases», última viñeta).
- [x] b. Fuera de alcance: las diez viñetas literales del hito, generadas desde `docs/ROADMAP.md` con un guion y comprobadas línea a línea contra él, y lo no especificado.
- [x] c. Clarificaciones: `## Clarifications` lleva una línea por pregunta de `gates/clarify-respuestas.json` (Q1 a Q5, con su criterio y su fuente; cuatro conservadoras), integradas en FR-017, FR-021, FR-024, FR-040, FR-041, FR-044, FR-061, FR-063, FR-067, FR-068, FR-074, US3, US6, «Edge Cases», SC-005, «Fuera de alcance» y «Relación con otros hitos».
- [x] d. Aceptación: «Criterios del hito, literales» lleva Entrega, Controles y Aceptación transcritos con un guion desde `docs/ROADMAP.md` (15 líneas, comprobadas línea a línea); SC-001 a SC-009.
- [x] e. Decisiones cerradas: el CENDOJ solo para resolver una resolución identificada, por el formulario, a sus metadatos (FR-006, FR-020, FR-021); GET y HEAD salvo el formulario declarado (FR-030, FR-031); el sobre (FR-010); el contrato de resultados por filas (FR-012, FR-016, FR-017); el grafo con procedencia y sin texto (FR-042); skills sin código y con las dos formas en la tabla (FR-060); el juez, nunca sin medir (FR-076); y la integración continua sin consultar el CENDOJ (FR-091).
- [x] f. Checklist veraz: cada marca cita dónde se cumple; con las clarificaciones integradas no queda ningún ítem sin marcar. El recuento de escenarios y los ids citados se volvieron a contar con `grep` y `awk` tras la corrección de la ronda 1.
- [x] g. Sin implementación: las formas concretas son del plan; los nombres que aparecen los da el hito.
- [x] h. Derivable: las formas de las referencias con sus ejemplos (FR-003 a FR-005), el código de cada resultado (FR-010 a FR-019), la clave y la vigencia de la caché (FR-040), el orden de los intentos de la skill (FR-061), lo que hace cita y lo que hace línea (FR-062, FR-063, FR-071) y lo que cuenta como entregado (FR-081). Lo que tenía dos lecturas con resultado distinto (FR-024, FR-041 y FR-067) quedó resuelto en `## Clarifications`. Corregido tras el rechazo del juez (ronda 1): FR-012, con US2 escenario 2 y su viñeta de Assumptions, ya no pide a la vez nombrar la referencia y callar el ROJ, y la misma salvedad —nada de la resolución encontrada que no esté en lo pedido— está en FR-041, FR-044, US6 escenario 6 y SC-005; FR-016 y FR-022 dan fila a todo estado distinto de 200 —5xx, la de FR-017 (4); cualquier otro, respuesta que no se reconoce (5)— y dicen que el «no encontrado» (3) nunca sale de un estado; FR-021 y FR-033, que una redirección de la página o del formulario no se sigue ni reenvía el formulario; y los requisitos de otros hitos se citan sin guion y con su hito (H2 FR 015, H4 FR 032, H21 FR 014), de modo que ningún id con guion queda sin definir en el spec.
- [x] i. Uso: «Uso, de fuera adentro», con quién pide cada salida y cómo —la skill lanza sus consultas de una en una en los dos modos (FR-061)—, su tamaño con su cálculo, el tiempo de una consulta por órdenes y por herramienta, y cuándo deja de darse cada señal. Corregido tras el rechazo del juez (ronda 2): FR-024 declara que las llamadas simultáneas de herramienta hacen cola en el ritmo compartido, cada una con su plazo, y que la que no alcanza turno termina con 4; y el párrafo de tiempo ya no dice que el ritmo no suma espera entre invocaciones más que por órdenes, y lleva el cálculo del modo de herramientas: de una en una, como mucho unos 15 s por llamada; cinco a la vez, seis turnos en 30 s para entre 10 y 15 peticiones.
- [x] j. Umbral con control: `cita_sin_resolver` y `sin_activar` nombran el job en rojo (FR-080, FR-082); las 300 líneas, `skills-check` (FR-060); los 30 días, las 10 resoluciones y el único envío por consulta, su test de `make ci` (FR-040, FR-014, FR-021); el ritmo, el test de la fila (FR-023); los cuatro umbrales del informe, FR-083; SC-001 a SC-007 y SC-009 llevan **Control**; y FR-084 fija las cuatro filas de «Controles de umbral». «Resume» no es un requisito con umbral: el hito lo deja sin control y lo anota como supuesto (FR-085).

## Notes

- Las tres marcas [NEEDS CLARIFICATION] del spec inicial eran las ambigüedades de alcance que el modo autónomo manda dejar al paso de clarify; sus respuestas, con su criterio y su fuente, están en `## Clarifications` del spec. Las demás lecturas que el spec fija salen del hito, del ADR 0036 o de la constitución, y están en Assumptions con la alternativa rechazada.
- En esta sesión no se ha ejecutado ningún test ni `make ci`, ni se ha consultado ninguna fuente. Lo comprobado está en la viñeta «Lo comprobado en esta sesión» de Assumptions.
- Corrección de la ronda 1 (corrector del spec): tampoco se ha ejecutado ningún test ni `make ci`. Lo ejecutado son búsquedas en `spec.md` —ningún id con guion sin definir, 81 ids definidos sin duplicados, 5 líneas de clarificación, 44 escenarios— y la lectura de `internal/httpx/cliente.go` (`seguirLaCadena`, `entregar`) y de `internal/httpx/errores.go` (`errorDe429`) para saber qué hace hoy el módulo con una redirección y con cada estado.
- Corrección de la ronda 2 (corrector del spec): tampoco se ha ejecutado ningún test ni `make ci`, y no se ha medido cómo reparte sus llamadas la app de escritorio. Lo ejecutado son las mismas búsquedas en `spec.md` tras la corrección —81 ids definidos sin duplicados, ninguno con guion sin definir, 5 líneas de clarificación, 44 escenarios— y la lectura de `internal/httpx/ritmo.go` (un turno por intervalo y por sitio, compartido por los clientes de un proceso; `sinTurno`), de `internal/httpx/robots.go`, de `internal/app/boe.go` y de `internal/cli/globales.go` (30 s de plazo por omisión), de las que sale el cálculo del modo de herramientas de «Uso, de fuera adentro».
