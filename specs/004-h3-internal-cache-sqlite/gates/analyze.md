# Specification Analysis Report — H3 · `internal/cache`: SQLite con TTL y `--offline`

**Modo**: desatendido (sin preguntas, sin remediación diferida). Artefactos analizados: `spec.md`,
`plan.md`, `tasks.md`, `data-model.md`, `contracts/*.md`, `checklists/requirements.md`,
`.specify/memory/constitution.md`.

## Hallazgos

| ID | Categoría | Severidad | Ubicación(es) | Resumen | Recomendación |
|----|----------|----------|-------------|---------|----------------|
| F1 | Inconsistencia | MEDIUM | tasks.md:55, tasks.md:57-67, tasks.md:265 | El texto dice «las siete excepciones» / «Las otras siete no añaden producto nuevo», pero enumera ocho tareas (T001, T006, T007, T009, T011, T012, T013, T014); la tabla de la rúbrica del juez (criterio `c. rebanadas_verdes`, línea 265) reproduce el error: lista solo siete IDs y **omite T013**, aunque el cuerpo del texto (línea 63, «T012 y T013 son validaciones…») sí describe T013 como excepción. | Corregir «siete» → «ocho» en ambas apariciones (líneas 55 y 57) y añadir T013 a la lista de la línea 265, para que la cuenta, la enumeración en prosa y la fila de la rúbrica coincidan. |
| F2 | Inconsistencia | MEDIUM | contracts/errores-y-codigos.md:40 (tabla §3, filas 1-15) | La frase introductoria de la tabla cerrada dice «Las **ocho** situaciones de SC-011 (marcadas ★) y las siete que el diseño hace inevitables» (8+7=15), pero el recuento real de marcas en la tabla es **nueve** filas con ★ (1, 2, 4, 5, 6, 8, 9, 10, 11) y **seis** sin marca (3, 7, 12, 13, 14, 15), es decir 9+6=15. La discrepancia surge porque el bullet 3 de SC-011 («ruta o variable inservibles») agrupa tres filas del contrato (4, 5 y 6) y sus bullets 4 y 5 («ausencia… en solo lectura» y «directorio inexistente… en solo lectura») se resuelven ambos en la misma fila 8; el recuento de filas ★ no es igual al recuento de bullets narrativos de SC-011, y el texto no lo advierte. | Cambiar la frase a algo que refleje la relación real, p. ej. «las ocho situaciones de SC-011 (que corresponden a las nueve filas marcadas ★, porque un bullet de SC-011 puede agrupar varias filas) y las seis que el diseño hace inevitables», o renumerar las marcas para que coincidan 8↔8. |

## Coverage Summary Table

Los 47 requisitos funcionales (FR-001 a FR-047) y los 14 criterios de éxito (SC-001 a SC-014) están
mapeados en la tabla «Trazabilidad: requisito → tarea» de `tasks.md` contra T001-T014. Verificación
exhaustiva confirma que:

- Ninguna FR ni SC queda sin al menos una tarea asociada.
- Ninguna tarea introduce trabajo no pedido por el spec (cada línea de tarea remite a requisitos y/o
  decisiones de `research.md`).
- Las tres decisiones de `Complexity Tracking` (dos clases de fallo no enumeradas en FR-033, la
  reapertura `immutable=1`, `ConReloj` exportado) están declaradas con alternativa rechazada y no
  aparecen como huecos de cobertura.

| Requirement Key | Has Task? | Task IDs | Notes |
|---|---|---|---|
| FR-001 a FR-047 (bloques completos) | Sí | T001–T014 (ver tabla de trazabilidad de tasks.md) | Cobertura 100 %; sin huecos detectados |
| SC-001 a SC-014 | Sí | T007, T008, T005, T003, T004, T006, T009, T010, T011, T012, T013 (según fila) | Cobertura 100 %; sin huecos detectados |

## Constitution Alignment Issues

Ninguno. `plan.md` §«Constitution Check» evalúa los nueve principios (I-IX) y las cinco reglas de
dependencia con veredicto y evidencia mecánica (subpruebas de `TestArquitectura`, listas de `depguard`,
`forbidigo`); la re-evaluación tras la fase de diseño también resuelve en «PASA» sin excepciones. No se
detectó ninguna afirmación del plan o de las tareas que contradiga un principio MUST de
`.specify/memory/constitution.md`.

## Unmapped Tasks

Ninguna. T001-T014 aparecen todas en la tabla de trazabilidad de `tasks.md` o en la sección
«Definition of Done: qué punto cubre qué tarea».

## Metrics

- **Total Requirements**: 47 FR + 14 SC = 61
- **Total Tasks**: 14
- **Coverage %**: 100 % (61/61 con al menos una tarea)
- **Ambiguity Count**: 0
- **Duplication Count**: 0
- **Critical Issues Count**: 0
- **High Issues Count**: 0
- **Medium Issues Count**: 2 (F1, F2)
- **Low Issues Count**: 0

## Next Actions

Ningún hallazgo es CRITICAL ni bloquea `/speckit-implement`: los dos hallazgos son defectos de conteo en
prosa dentro de documentos de trazabilidad/autocomprobación (`tasks.md` y el contrato de errores), no
afectan a ningún requisito funcional, código de salida, regla de arquitectura ni control mecánico —las
clasificaciones situación→clase→código y la delimitación de las siete/ocho tareas restantes ya son
correctas en la práctica (T013 sí se ejecuta y sí es una excepción; las 15 filas de la tabla de errores
son correctas una a una). Recomendación: corregir ambas frases de conteo directamente en `tasks.md` y en
`contracts/errores-y-codigos.md` antes o durante la primera tarea que toque esos ficheros, sin bloquear
`/speckit-implement`.

Dado el modo desatendido, no se abre ronda de preguntas ni se ofrece un plan de remediación diferido:
las dos correcciones anteriores son la remediación completa y quedan a criterio de quien ejecute
`/speckit-implement` aplicarlas como ajuste editorial de bajo riesgo.
