# Specification Analysis Report — H1 · Kernel CLI: multicall, flags globales, exit codes, sobre de salida

**Modo**: desatendido (sin preguntas, sin remediaciones pospuestas). Artefactos analizados: `spec.md`,
`plan.md`, `tasks.md`, más `data-model.md`, `contracts/*.md`, `research.md` y `checklists/requirements.md`
como contexto de apoyo, y `.specify/memory/constitution.md` como autoridad no negociable.

## Hallazgos

Ningún hallazgo ha sobrevivido la verificación. La tríada spec → plan → tasks es internamente consistente:
las 60 FR y las 15 SC tienen al menos una tarea que las cubre, las 27 decisiones de `research.md` (D1–D27)
citadas desde `plan.md` existen todas, el `data-model.md` refleja exactamente las cinco *Key Entities* del
`spec.md`, los 12 escenarios de `quickstart.md` coinciden con los que `tasks.md` (T020) dice ejecutar, las
5 dependencias nuevas están todas en la lista cerrada de la constitución §V, y el único punto que podría
leerse como una reapertura de una decisión cerrada de `CLAUDE.md` — el código de salida **1** para el
fallo «inesperado» — está explícitamente justificado en `plan.md` (Constitution Check, re-evaluación tras
la fase 1) y en `contracts/banderas-y-exit-codes.md:86-89`: los códigos 0, 2, 3, 4, 5 y 6 no se tocan, y el
1 llena un hueco que `CLAUDE.md` no cubría, no lo reabre. El checklist `checklists/requirements.md` tiene
sus 16 ítems marcados `[x]` y ninguna marca pendiente.

No hay fila que reportar: tabla vacía.

| ID | Category | Severity | Location(s) | Summary | Recommendation |
|----|----------|----------|-------------|---------|----------------|
| — | — | — | — | Sin hallazgos | — |

## Resumen de cobertura

**Requisitos funcionales (FR-001 … FR-060):** las 60 tienen al menos una tarea. Mapeo completo en
`tasks.md` §«Requisitos → tareas»; no se repite aquí por brevedad. Ninguna FR queda sin tarea; ninguna
tarea queda sin FR o SC que la justifique (T013 y T017 tocan controles del hito literal en vez de una FR
individual, y están citadas explícitamente en «Criterios del hito, literales» y en la Definition of Done).

**Criterios de éxito (SC-001 … SC-015):** las 15 tienen al menos una tarea, según `tasks.md` §«Criterios de
éxito → tareas». Las que exigen trabajo construible (cobertura, e2e, reglas de arquitectura, reproducibilidad
de huella) están cubiertas por T001, T002, T006, T014, T015, T016, T017 y T020.

**Historias de usuario (US1 … US6):** las 6 tienen prueba independiente declarada en el spec y tareas
asociadas en `tasks.md` §«Historias de usuario → tareas». Ninguna historia queda sin tarea ni sin prueba
independiente.

**Edge Cases del spec (11 bullets):** los 11 tienen respaldo explícito en una FR y en una tarea — precedencia
del multicall (FR-004/T010), enlace no registrado (FR-005/T010), invocación sin applet deducible (FR-006/T010),
`--timeout` inválido (FR-020/T005), banderas excluyentes (FR-028, FR-042, FR-049/T005, T007), banderas sin
objeto todavía (FR-021, FR-023, FR-024/T005), `data` no serializable (`data-model.md` §1/T001, T006), tubería
cerrada (T008 `TestEscrituraFallida`), error con contenido parcial (FR-045/T006), registro duplicado o en
colisión (FR-008/T009). El único bullet sin control mecánico propio es el de stderr redirigido a stdout por
quien invoca: es una limitación reconocida y no una garantía que el binario pueda dar, correctamente descrita
como tal y no como un requisito pendiente.

## Alineación con la constitución

Sin problemas. Los siete principios están evaluados uno a uno en `plan.md` §«Constitution Check» con
veredicto `✅ Cumple` y evidencia concreta (rutas, líneas de código de terceros verificadas, reglas de
`depguard`/`forbidigo`). Las cinco reglas de dependencia de `docs/ROADMAP.md` §2 están todas activas
—incluidas las que quedan «activas y vacías» porque su paquete objetivo no existe todavía— con doble capa
de comprobación (lint + `internal/arch_test.go`), tal como exige FR-051. Ninguna dependencia nueva sale de
la lista cerrada de la constitución §V. Ninguna decisión cerrada de `CLAUDE.md` se reabre.

## Tareas sin mapear

Ninguna. Las 21 tareas (T001–T021) están todas referenciadas desde al menos una FR, SC o historia de
usuario, o desde la Definition of Done (`docs/ROADMAP.md` §1) para las tareas de cierre (T017–T021).

## Métricas

- Requisitos funcionales totales: 60 (FR-001–FR-060)
- Criterios de éxito totales: 15 (SC-001–SC-015)
- Historias de usuario: 6
- Tareas totales: 21 (T001–T021)
- Cobertura de requisitos (≥1 tarea): 100 %
- Cobertura de criterios de éxito (≥1 tarea): 100 %
- Tareas sin requisito ni criterio asociado: 0
- Ambigüedades detectadas: 0
- Duplicaciones detectadas: 0
- Incidencias CRITICAL: 0
- Incidencias HIGH: 0
- Incidencias MEDIUM: 0
- Incidencias LOW: 0

## Próximas acciones

No hay incidencias CRITICAL ni HIGH que resolver. El artefacto puede pasar a `/speckit-implement` sin
cambios en `spec.md`, `plan.md` ni `tasks.md`. Modo desatendido: no se ofrece plan de remediación pospuesto
porque no hay nada que remediar.
