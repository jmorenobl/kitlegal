# Specification Analysis Report — H25 · El juez de `jurisprudencia`

**Feature**: `specs/020-h25-el-juez-de` · **Artefactos leídos**: `spec.md` (466 líneas), `plan.md` (405), `tasks.md` (385),
`quickstart.md`, más `research.md` y los contratos solo por sus apartados citados · **Constitución**: 2.13.0.
**Modo**: autónomo y de solo lectura (únicamente se escribe este informe). No se ha ejecutado ningún test ni `make ci`:
todo lo que sigue es lectura del árbol y aritmética hecha a mano, y lo que no se ha podido medir se dice.

## Resultado

Sin hallazgos CRITICAL ni HIGH. Seis hallazgos LOW o MEDIUM, ninguno bloquea `/speckit-implement`.

## Specification Analysis Report

| ID | Category | Severity | Location(s) | Summary | Recommendation |
|----|----------|----------|-------------|---------|----------------|
| F1 | Inconsistency | MEDIUM | `internal/evals/umbrales.go:183-190` (comentario de `umbralesDelInforme`); `tasks.md` T004 y T007 | El comentario dice que las evals de `jurisprudencia`, «que no tiene juez ni objetivo, son cuatro» umbrales. Con H25 son doce, diez que deciden. T004 declara `umbrales.go` solo para una corrección condicional y no manda actualizar ese comentario; T007 solo declara `doc.go` (que sí cita el caso en su línea 158). Lo dejaría falso a sabiendas, la misma clase que el barrido y el criterio f de la revisión final marcan. | Añadir `internal/evals/umbrales.go` a las rutas de T007 y a su texto («el comentario de `umbralesDelInforme`»), o mandarlo en T004, que ya lo declara. |
| F2 | Coverage / Underspecification | MEDIUM | `spec.md` FR-073, Assumptions (penúltimo de «límite de ritmo»); `plan.md` research S2; `tasks.md` «Estrategia» | Nueve sesiones a la vez (4 + 4 + 1) contra la misma suscripción no se han medido, y el spec decide que bajar la concurrencia no es una corrección del run. Si el límite de ritmo corta sesiones, el cierre queda en `fallo` por «sin medir» sin un camino de reparación previsto dentro del run; queda para la persona en el informe final. Es una decisión consciente y registrada, no un defecto de redacción, pero es el riesgo de ejecución más probable del hito. | Sin cambio de artefactos (la lectura conservadora está registrada). Que el informe final destaque este riesgo por separado de las marcas del juez. |
| F3 | Inconsistency | LOW | `quickstart.md` §10, párrafo «Esperado»; `spec.md` FR-021 y SC-001 | §10 espera los doce umbrales «todos con `cumple` verdadero». Pero `afirma_que_existe` lleva `umbral` 0 y su `cumple` solo informa (FR-021): con una respuesta que diga que una sentencia existe, `cumple` es falso y SC-001 se sostiene igual (solo pide que «se publique con su recuento»). | En §10, decir «los diez que deciden con `cumple` verdadero» y que los dos de `afirma_que_existe` se publican con su recuento. |
| F4 | Inconsistency | LOW | `plan.md` «Tests existentes que cambian» y «Source Code»; `tasks.md` cabecera y T003 | `TestJuzgarSentencias` (`internal/evals/sentencias_test.go:130`) y sus ayudantes de seis evals (`require.Len(…, 6)`) los añade tasks.md a T003 «comprobado en el árbol», pero el plan no los nombra y su árbol de fuentes no lista `sentencias_test.go`. Los controles `ci:` no se ven afectados. | Aceptable tal cual, porque tasks.md lo declara y T003 lleva la ruta. Si se regenera el plan, añadirlo. |
| F5 | Inconsistency | LOW | `spec.md` SC-003 y FR-103 («2 de 2 mutaciones»); `tasks.md` T004 (tres mutaciones) y tabla de controles («2 de 2») | El spec cuenta dos mutaciones («de una en una» en `evals` y en `medida`); T004 mantiene tres, con «sin la skill en `medida`», que cubre el otro mandato de FR-103 («lleva las dos skills con juez»). La tabla de controles de tasks.md dice «2 de 2» y T003/T004 reparten las mutaciones de forma distinta a como las narra el plan. | Sin cambio funcional. Si se toca, decir «2 mutaciones de concurrencia y 1 de matriz». |
| F6 | Coverage | LOW | `tasks.md` T004 | T004 reúne 35 requisitos, 10 rutas y las cuatro copias de la carpeta del juez. Está justificada (la carpeta sola deja cuatro tests en rojo, research M4), pero es la tarea de mayor riesgo de cuarentena del hito y de ella dependen T005 y toda la aceptación. | Sin cambio: la alternativa de partirla dejaría `make ci` en rojo entre tareas, que el plan prohíbe. |

## Coverage Summary

| Requisito | ¿Tiene tarea? | Tareas | Notas |
|---|---|---|---|
| FR-001 a FR-004 (carpeta, clases, rúbrica) | Sí | T004 (T008 constata) | `clases.yaml` sin `cuenta_su_proceso` |
| FR-005 (nada cambia en `boe-legislacion` / `legal-core`) | Sí | T001 a T005, T008 | Restricción de todas las tareas |
| FR-010 a FR-012 (qué recibe el juez) | Sí | T004 | Los textos de la sesión no cambian `sesion.go`: leído genérico por `mcp__<servidor>__` |
| FR-013 (`sentencia`) | Sí | T001, T004 | |
| FR-020 a FR-025 (umbrales y veredicto) | Sí | T004 | FR-024 también T003 y T007 (`doc.go`) |
| FR-026 | Sí | Restricción de todas las tareas, T005, T008 | |
| FR-027 (reparaciones del cierre) | Por diseño | T008 constata los apartados | Lo ejerce el workflow, no una tarea |
| FR-030 a FR-033 (medida versionada) | Sí | T004 (T008 constata) | |
| FR-040 a FR-045 (reconstrucción y derivaciones) | Sí | T002, T004, T005 | `registroDeLaSesion` (`medida.go:884`) solo registra `boe` y `graph` hoy: T002 lo amplía, coherente con el plan |
| FR-050, FR-051 (medida de las dos skills) | Sí | T004, T005 | `evals.yml:376` solo lleva `boe-legislacion` en la matriz: T004 |
| FR-060 a FR-066 (cuatro evals) | Sí | T003 | |
| FR-070 a FR-072 (concurrencia y topes) | Sí | T003, T004, T007 (comentarios) | Aritmética rehecha a mano en esta sesión, ver abajo |
| FR-073 | Por diseño | T008 constata | Lo mide el cierre |
| FR-080 a FR-084 (`SKILL.md` v0.1) | Sí | T006, T007 (FR-084) | `git apply --check` del diff del contrato: correcto en esta sesión |
| FR-090 a FR-093 (documentación) | Sí | T007 | Cada afirmación falsa de FR-093 existe hoy en `CONTRIBUTING.md` (líneas 625, 703-706, 942, 955-956) |
| FR-095, FR-096 | Sí | Restricción de todas, T007, T008 | |
| FR-100 a FR-112 (controles) | Sí | T001 a T008 | De los 24 nombres de test buscados, 21 existen y se han localizado con `^func Test…\(`; los 3 que faltan (`TestTextoQuitado`, `TestPreguntasDelSondeo`, `TestReconstruccionDeJurisprudencia`) los crea una tarea |
| SC-001 | Por diseño | T004 construye los controles | Lo mide el cierre del workflow |
| SC-002 a SC-013, SC-015 | Sí | T001 a T007 | |
| SC-014 | No (humana) | — | Después del run, por la persona |

### Verificaciones hechas en el árbol en esta sesión

- `skills/jurisprudencia/SKILL.md`: 195 líneas, una sola línea con «CAPTCHA» (la 19); el diff del contrato tiene 9 líneas añadidas y 7 quitadas y `git apply --check` pasa, así que 195 + 9 − 7 = 197, como dicen el plan y T006.
- `evals/jurisprudencia/` tiene hoy seis evals (`01` a `06`) y `evals/boe-legislacion/juez/` cinco ficheros; `evidencias/adr-0037-jurisprudencia/` lleva la rúbrica, el esquema, los casos, la medida, las preguntas y los dos informes de sondeo.
- Peores casos con la fórmula de `CONTRIBUTING.md` (§ del job): sesiones, 485 + (⌈61/4⌉ + ⌈60/4⌉) × 272 = 8 917 s (el intermedio de T003); juez, 60 + (⌈30·6/4⌉ × 2) × 40 = 3 660 s; total 12 577 s (209,6 min < 352). Medida: 485 + 60 + ⌈249·6/4⌉ × 40 = 15 505 s (258,4 min < 269). De una en una: 33 397 s de sesiones y 47 857 s con el juez. Todas coinciden con spec y plan; son un cálculo a mano, no una ejecución de `TestDefinicionDelJob`.
- `evals.yml` ya instala el Claude Code del juez sin condicionarlo a una skill concreta (línea 251), y su `concurrencia` de `jurisprudencia` es hoy 1 (línea 180).
- Ninguna marca `TODO`, `TKTK`, `???`, `NEEDS CLARIFICATION` ni `<placeholder>` en los seis documentos (las dos menciones de «TODO» son la prohibición de dejarlos).

## Constitution Alignment Issues

Ninguno. Comprobado contra la constitución 2.13.0:

- **III (e2e primero)**: la desviación «Aceptación e2e: no aplica» está en Complexity Tracking con su motivo y su precedente (H24), y tasks.md no lleva tarea `[aceptacion]` en coherencia con ella.
- **«Umbrales que se cumplen» (ADR 0029/0037)**: los seis umbrales que deciden y las filas medidas en `make ci` tienen control en «Controles de umbral», con tarea y test que lo ve fallar; ninguna tarea los rebaja (FR-026 en la batería de cada tarea).
- **Reglas del modo desatendido (ADR 0032)**: ninguna tarea usa red, abre sesión con modelo ni ejecuta `make evals*`, `TestSondeo` o los guiones de la evidencia; ninguna escribe en `evidencias/` ni la nombra por ruta.
- **Proporcionalidad / convergencia**: cada mecanismo del plan se traza a un FR («Trazabilidad»). La cláusula FR-027 reproduce las cuatro condiciones de la clarificación (cuándo, qué, traza, quién juzga) y el tope de tres mediciones y dos reparaciones.

## Unmapped Tasks

Ninguna. T008 (cierre de la Definition of Done) no cambia producto y se traza a FR-001, FR-005, FR-011, FR-026, FR-027, FR-030, FR-033, FR-073, FR-083, FR-095, FR-096, FR-100, FR-101 y SC-012.

## Metrics

- Requisitos funcionales: 64 (FR-001 a FR-112, con huecos de numeración deliberados) · criterios de éxito: 15 · **total 79**.
- Tareas: 8 (T001 a T008), ninguna `[P]`, `[aceptacion]` ni `[datos]`.
- **Cobertura**: 64 de 64 FR con tarea o con el mecanismo declarado (FR-027 y FR-073 por el workflow y la constatación de T008); 13 de 15 SC con tarea, y los otros 2 por diseño (SC-001, que mide el cierre, y SC-014, humano).
- Ambigüedades abiertas: 0 · duplicados: 0 · **CRITICAL: 0** · HIGH: 0 · MEDIUM: 2 (F1, F2) · LOW: 4 (F3, F4, F5, F6).

## Next Actions

- Sin CRITICAL: se puede ejecutar `/speckit-implement`.
- Recomendado antes (o dentro de T007): declarar `internal/evals/umbrales.go` para corregir el comentario de F1.
- F2 es un riesgo de ejecución que el run no puede cerrar por su cuenta: conviene que el informe final lo muestre aparte de las marcas del juez.
- F3 a F6 no requieren cambio para implementar.

Los ganchos opcionales `before_analyze` y `after_analyze` de `.specify/extensions.yml` (commit automático) no se han ejecutado: esta sesión es de un paso y solo escribe este fichero; el commit lo hace el workflow.
