# Specification Quality Checklist: H36 · El motor de cómputo descrito por datos (ADR 0061)

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-11-22
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs) — el spec nombra `LeerTipos`, `plazos.Calcular`,
  `CodigosDeAviso`, `EtiquetaDeAviso`, `go/ast` y los ficheros de test porque el hito los nombra como contrato o como
  control (FR-009, FR-010, FR-013, FR-017, FR-023); no fija tipos de Go, firmas, paquetes nuevos ni cómo se juntan el
  cómputo real y los tipos de prueba, que son del plan. Las claves de `data/plazos.yaml` (§C.1) las fija porque el hito
  se lo pide («Los nombres de las claves los fija el spec»).
- [x] Focused on user value and business needs — US1 (la skill recibe lo mismo), US2 (un cómputo de otra materia entra
  como datos), US3 (sin tocar Go), US4 (el esquema del agente no cambia con cada paquete).
- [x] Written for non-technical stakeholders — cada historia empieza por quién lo ve: la skill, quien escribe un
  paquete, el agente que conecta.
- [x] All mandatory sections completed — escenarios, requisitos, criterios de éxito, supuestos y «Fuera de alcance».

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain — ninguno: el hito y el ADR 0061 fijan ejes, reglas, error nuevo, cómputo
  civil, casos y controles; las claves y los textos que el hito deja al spec están en §C.1, §C.6 y §C.7, con su
  motivo en «Assumptions».
- [x] Requirements are testable and unambiguous — cada FR termina en «Comprobable:» con su escenario, su caso (C1 a
  C10, E1 a E13) o su control; los mensajes nuevos (FR-006, FR-007) y la ayuda (FR-012) van byte a byte en §C.6 y §C.7;
  la regla 4 de §C.2 precisa el caso de un cómputo sin tipos.
- [x] Success criteria are measurable — SC-001 a SC-009 dan recuentos (10 casos, 13 casos, 3 juegos de datos, 5
  preceptos, 0 diferencias, 0 apariciones) y cada uno nombra la comprobación de `make ci` o del cierre que sale en rojo.
- [x] Success criteria are technology-agnostic (no implementation details) — se miden sobre lo que ve quien usa el
  binario o escribe los datos (fecha, avisos, código, mensaje, esquema); los controles nombran el test que los hace
  cumplir, como pide el criterio j.
- [x] All acceptance scenarios are defined — US1 a US6 cubren la aceptación y cada control del hito (mapa en
  «Criterios del hito (literales)»).
- [x] Edge cases are identified — periodo que cruza el año, festivo dentro de un periodo `inhabil`, festivo de una parte
  en él, varias pasadas por un `aviso`, `dia-siguiente` en inhábil, inicio dentro de un periodo, datos que no cargan
  con la regla genérica.
- [x] Scope is clearly bounded — «Fuera de alcance» con toda la lista cerrada del hito y lo que rechaza el ADR 0061.
- [x] Dependencies and assumptions identified — «Assumptions» (gramática del código, textos de municipio, `da`
  obligatoria, la fila «no cargan», las referencias del paquete, la grabación); el hito no tiene dependencias
  pendientes.

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria — FR-001 a FR-025, cada uno con su escenario, su caso
  o su SC.
- [x] User scenarios cover primary flows — lo de hoy igual (US1), el civil (US2), el ensayo (US3), el esquema (US4),
  `LeerTipos` (US5), lo grabado (US6).
- [x] Feature meets measurable outcomes defined in Success Criteria — SC-001, SC-002, SC-003 y SC-005 son la
  aceptación literal del hito.
- [x] No implementation details leak into specification — ver la primera nota de «Content Quality».

## Notes

- Los requisitos de H9 se citan sin guion («FR 007 de H9») para que el precheck no los tome por requisitos de este spec
  sin definir.
- La sección «Quién consume cada salida, cuánto y hasta cuándo» responde al criterio i: el `data` no cambia (de 0,77 a
  0,86 KB por plazo, una o dos llamadas por pregunta, sin depender del uso acumulado); el esquema de la herramienta,
  una vez por conexión, ya no crece con los datos; el aviso de periodo, una línea solo con el tramo que lo toca.
- Los únicos umbrales numéricos son recuentos de casos y de apariciones; cada SC nombra la comprobación que sale en rojo
  (criterio j). Las cifras de tamaño de «Quién consume» son medidas o estimaciones, no umbrales.
