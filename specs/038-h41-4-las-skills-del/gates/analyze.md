# Specification Analysis Report — H41.4 (`038-h41-4-las-skills-del`)

Análisis de solo lectura de `spec.md`, `plan.md` y `tasks.md` (con `contracts/` y `research.md` solo donde hizo falta).

**Qué he medido en esta sesión** (órdenes ejecutadas y vistas terminar): `wc -m` de los cuatro textos de
`laboratorio/producto-descripciones-cortas/` (418, 421, 427 y 428 con su salto de línea: 417, 420, 426 y 427 sin él;
suma 1694 − 4 = 1690); que `ab951c22` resuelve a `ab951c227e1fc8acac8bd6b23534e275e4a36a76`; que
`skills/consulta-juridica/` en esa base es un solo fichero (`SKILL.md`); que la SHA-256 del `SKILL.md` de la base sin las
líneas de la clave `description` es `bd089909060d0fd82c1c6d2ee4ae148f6f0a1b2b7d7eefd476d164dc610a205a`, la que fija
`contracts/controles.md` §2.1 (calculada con un `awk` equivalente a `descripcionYCuerpo`); que existen
`maximoDeUnaDescripcionCorta = 450`, `maximoDeLasDescripcionesCortas = 1700`, `descripcionesCortas`,
`descripcionYCuerpo`, `leerLosFicheros` y los tests que el plan declara «sin cambiar»; que `descripciones_del_producto_test.go`
aún no existe (lo esperado antes de implementar) y que no hay marcadores `TODO`/`???`/`NEEDS CLARIFICATION` en los tres
documentos.

**Qué no he podido medir**: que los cuatro bloques `>-` de `contracts/skills.md` §2 se plieguen byte a byte a los `.txt`
(un guion de un paso fue rechazado por el sandbox, que pedía aprobación; lo afirma `research.md` V2, que no he repetido);
que Claude Code lea el bloque `>-` como el escalar del sondeo de H41.3 (S1 del research, solo lo mide el sondeo); y las
cifras de H41.3 (3313 caracteres de hoy, 640 sesiones) que el spec ya declara «no medidas en esta sesión». No he releído
`.specify/memory/constitution.md` entero: el contraste con la constitución se apoya en la tabla «Constitution Check» del
plan y en las reglas que `CLAUDE.md` resume.

## Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| F1 | Inconsistencia | LOW | plan.md «Scale/Scope» y «Diseño» (`TestDescripcionesDelProducto`) · contracts/controles.md §1.3 · tasks.md T003 | La tabla §1.3 tiene 9 filas: «la rama» y 8 mutantes. Tasks T003 dice «nueve casos», que es correcto; el plan dice «nueve mutantes sobre copias» y «16 subtests (10 y 6)», que serían 15 (9 y 6). `TestConsultaJuridicaDesdeLaBase` sí cuadra: 6 incluyendo «la rama» | Corregir en el plan «ocho mutantes» y «15 subtests (9 y 6)». No afecta a la implementación: T003 y el contrato coinciden |
| F2 | Cobertura | LOW | spec.md FR-006 (frase final) | «`INSTALACION=producto-descripciones-cortas` sigue aceptándose y pasa a instalar las mismas descripciones que `INSTALACION=producto`» es una consecuencia declarada, sin control ni tarea que la compruebe. Los controles de FR-006 (`TestDescripcionesCortas`, `TestVariantesDeControl`) no la miran | Aceptable: sale de que `skills/` y la variante ya coinciden (lo comprueba T003) y no cambia el arnés. Si se quisiera, bastaría decirlo en el informe final |
| F3 | Cobertura | LOW | spec.md FR-006, FR-007 · plan.md «Controles de umbral» (última nota) | «El diff no toca `laboratorio/`, `web/`, `mcp/`, el empaquetado ni `evals/`» no lo mide `make ci`: lo guarda el guardián de diff del workflow y lo enseña T005 (`git diff --name-only ab951c22`) | Ninguna: el plan lo declara. Queda como control del workflow, no de `make ci` |
| F4 | Cobertura | LOW | spec.md FR-008 · tasks.md T005 | La viñeta del `CHANGELOG.md` no tiene control automático (el spec lo excluye expresamente en «Fuera de alcance»). `CHANGELOG.md` no menciona aún H41.4 (consultado con `grep`) | Ninguna: T005 la escribe byte a byte desde `contracts/skills.md` §4 y la revisión la lee |
| F5 | Subespecificación | LOW | spec.md «Assumptions» / research S1 | Que el host lea el bloque `>-` como el escalar del sondeo de H41.3 no se puede comprobar sin modelo; la comprobación es el sondeo de la persona (FR-011, SC-006) | Ninguna: el riesgo está declarado y lo cubre la aceptación de la persona; el plan lo pasa al informe final |
| F6 | Ambigüedad | LOW | spec.md «Contexto» y «Uso, de fuera adentro» | Las cifras de partida (3313 caracteres de hoy, 640 sesiones, intervalos de Wilson) vienen del hito, del ADR 0055 y de H41.3, y el propio spec las marca «no medidas en esta sesión». Las cuatro medidas de la rama sí las he reproducido (417/420/426/427, 1690) | Ninguna |

Sin hallazgos CRITICAL ni HIGH ni MEDIUM.

## Resumen de cobertura

Requisitos del spec: FR-001 a FR-011 y SC-001 a SC-006. Los SC-001 a SC-004 son la medida de FR-001 a FR-009; SC-005 y
SC-006 son de FR-010 y FR-011.

| Clave | ¿Tiene tarea? | Tareas | Notas |
|-------|---------------|--------|-------|
| FR-001 (descripción = variante, byte a byte) · SC-001 | Sí | T001, T002, T003 | Guion congelado (8 `grep`), cambio en `skills/` y `TestDescripcionesDelProducto` con mutantes |
| FR-002 (bloque `>-` de más de una línea) | Sí | T001, T002, T003 | Mutantes del escalar y del bloque de una línea |
| FR-003 (nada más cambia) · SC-002 | Sí | T002, T004 | Tres familias por `TestSkillsDelProductoDesdeControl`; `consulta-juridica` por huellas de `ab951c22` (T004) |
| FR-004 (450 / 1700 / 900 / sin XML) · SC-003 | Sí | T002, T003 | 450 y 1700 en T003; 900 y XML por `TestSkillsDelRepositorio`, sin cambiar |
| FR-005 (único cambio en una suite congelada) | Sí | T002 | `TestSkillsDelProductoDesdeControl`, solo lo de la `description` |
| FR-006 (`laboratorio/` intacto) | Sí | T002, T005 (verificación) | Control: `TestDescripcionesCortas` y `TestVariantesDeControl` sin cambiar; el diff lo guarda el guardián (F3) |
| FR-007 (nada generado cambia) | Sí | T002, T005 (verificación) | `make skills-sync` + `git status --porcelain` |
| FR-008 (CHANGELOG) | Sí | T005 | Sin control en `make ci`, por el spec (F4) |
| FR-009 · SC-004 (suites en verde) | Sí | T002 a T005 | `make ci` tras cada tarea |
| FR-010 · SC-005 (evals y juez del cierre) | Sin tarea, por diseño | — | Lo mide el cierre del workflow; el plan nombra los umbrales por skill en «Controles de umbral» |
| FR-011 · SC-006 (sondeo de la persona) | Sin tarea, por diseño | — | El run no lo lanza (spec «Fuera de alcance»); el quickstart, escenario 8, da las órdenes |

**Historias**: US1 (T001–T003), US2 (T002, T004), US3 (T002, T005) con tareas; US4 sin ellas, a propósito (Phase 6).

## Alineación con la constitución

Sin conflictos. El plan trae la tabla de principios I a X y reglas R1 a R7 con su cumplimiento, y no hay nada que contradiga
las decisiones cerradas de `CLAUDE.md`: no cambia código de producción, no toca fuentes ni red, mantiene los límites del
ADR 0046, no rebaja ningún umbral (FR-010) y deja fusionar y el sondeo a la persona (ADR 0018, 0021, 0055). Test primero
(III): T001 escribe la suite de aceptación antes del cambio, y cada mutante nuevo se ve fallar (ADR 0029).

## Tareas sin requisito

Ninguna. T001 sirve a FR-001/FR-002 (guion); T002 a FR-001 a FR-005 y FR-009; T003 a FR-001, FR-002 y FR-004; T004 a
FR-003; T005 a FR-006 a FR-009.

## Orden y dependencias

Coherentes: T002 va antes de T003 (con las descripciones de hoy, `TestDescripcionesDelProducto` saldría en rojo); T002
junta las descripciones y el cambio de `TestSkillsDelProductoDesdeControl` porque por separado dejaría `make ci` en rojo
(el spec y el research V4 lo explican); T004 añade al fichero de T003 y reutiliza sus auxiliares; las rutas que declara
cada tarea cubren los ficheros que crea o modifica. Ninguna tarea usa red, modelo ni credenciales.

## Métricas

- Requisitos: 11 FR + 6 SC = 17 (los SC de medida se cubren con los FR que miden).
- Tareas: 5.
- Cobertura (FR de entrega con al menos una tarea): 9 de 9 (FR-001 a FR-009); FR-010 y FR-011 sin tarea por diseño,
  con su control o su responsable declarados.
- Ambigüedades: 0 abiertas (F6 es una cifra declarada no medida).
- Duplicaciones: 0.
- Hallazgos CRITICAL: 0 · HIGH: 0 · MEDIUM: 0 · LOW: 6.

## Siguientes acciones

Ningún hallazgo bloquea `/speckit-implement`. F1 es una errata de recuento en el plan que no cambia ninguna tarea ni
contrato; F2 a F6 son límites ya declarados en el spec y el plan. Lo que queda para la persona, como dice el plan: el
sondeo de las dos filas de 320 sesiones (FR-011), que el run no lanza, y S1.
