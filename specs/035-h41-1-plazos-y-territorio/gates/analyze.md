# Informe de análisis · H41.1 `plazos-y-territorio` manda leer sin condición

Artefactos analizados: `spec.md`, `plan.md`, `tasks.md` (con `contracts/`, `research.md`, `gates/supuestos.md` como apoyo) y la
constitución. Análisis de solo lectura; ningún fichero se ha modificado salvo este informe.

**Qué he medido y qué no.** He ejecutado `check-prerequisites.sh` (sale con los cinco documentos) y he leído el repositorio:
`h41-skills-instaladas.txtar` (líneas 8-55), `remisiones_test.go` (`fueraDeLasRemisiones`, `remisionesDelPlan`),
`docs/SKILLS.md` (líneas 142 y 160), `CHANGELOG.md` (líneas 64-84), los tamaños de `skills/plazos-y-territorio/` y el
`SKILL.md` de `-sin-condicion`. No he ejecutado `make ci` ni ningún test (el análisis no los necesita, y no hay nada que
probar antes de implementar). El resto de cifras del spec las doy por las que traen los artefactos.

## Specification Analysis Report

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| I1 | Inconsistencia | MEDIUM | spec.md FR-008, FR-012, SC-005; tasks.md T002; supuestos.md (3.ª línea) | T002 edita `internal/evals/laboratorio_sondeo_test.go` (`origenesDeLaVariante`, sale el segundo origen de `plazos-y-territorio-producto`). Ningún FR lo pide: FR-008 dice que el arnés y sus tests siguen «sin cambiar ningún valor esperado», FR-012 prohíbe cambiar nada fuera de FR-001 a FR-011 y FR-013, y SC-005 habla de cero bytes en «el arnés del laboratorio». Solo lo justifican research D12 y `supuestos.md`; la traza del plan lo pone bajo FR-005/FR-008. | Sin riesgo de ejecución (la ruta está declarada en T002 y el guardián la admite; sin el cambio, `make ci` sale en rojo). Está registrado como supuesto y en «Para la persona» del plan. Aceptable tal cual; si se reedita el spec, nombrarlo en FR-005 o en la lista de excepciones de FR-012 para que SC-005 no lo contradiga. |
| C1 | Cobertura | MEDIUM | spec.md «Fuera de alcance»; plan.md «Cierre y pendientes»; research D10 | Tras el hito quedan textos falsos sin tarea que los corrija: el ejemplo de remisión de `docs/SKILLS.md` §1 (viñeta de `municipios.md` de `plazos-y-territorio`) y `CONTRIBUTING.md` (líneas 1813-1814 y 1828-1830 según research D10, que no he releído), que seguirán diciendo que `skills/plazos-y-territorio` es su control byte a byte. | Decisión consciente por la lectura conservadora (FR-012) y ya llevada al informe final. No bloquea; la persona debe ver ambos pendientes al revisar. |
| A1 | Ambigüedad | LOW | spec.md FR-004, SC-002; tasks.md T002 (mutantes 10 881 / 10 931 / 11 001) | La cifra exacta de 10 882 caracteres no la he podido reproducir de forma independiente: contando el cuerpo de `-sin-condicion` (todo lo que sigue al cierre del frontmatter) con `awk` y `wc -m` salen 10 991, y M1 quita 110 caracteres (la frase de 105 más el espacio y el salto con sangría de 4 de las líneas 68-69), es decir 10 881. La diferencia de 1 depende de cómo mida el repo el cuerpo y de la localización de `wc`, así que puede no ser un defecto. | Sin acción previa. Lo confirma quickstart §4 («cuerpo 10 882») al implementar T002; si saliera otra cifra, los mensajes esperados de los mutantes («cuerpo de 11001 caracteres») cambiarían en la misma medida. |
| X1 | Inconsistencia | LOW | spec.md «Input» | El spec cita `docs/roadmap/in-progress/H41.1-…md` como la entrada, pero el fichero está en `docs/roadmap/done/` (commit `e041e042`; `in-progress/` solo tiene el README). | Cosa del workflow, no del contenido; ninguna tarea usa esa ruta. Nada que hacer. |
| U1 | Subespecificación | LOW | tasks.md T002 | T002 junta skill, referencia, dos ficheros de test, el guion de H41 y ocho mutantes. Está justificada (con la skill cambiada, cualquier parte que falte deja `make ci` en rojo; research D7, V5, V13), pero un fallo a mitad no tiene camino de vuelta descrito. | Aceptable: la tarea se rehace entera o va a cuarentena. Sin cambio. |

Sin hallazgos de duplicación ni de marcadores sin resolver (`TODO`, `TBD`, `[NEEDS CLARIFICATION]`: ninguno en spec, plan ni tasks).

## Coverage Summary

| Requisito | ¿Tarea? | Tareas | Notas |
|-----------|---------|--------|-------|
| FR-001 (órdenes de leer) | Sí | T001, T002 | Guion congelado y `TestSkillsDelProductoDesdeControl` |
| FR-002 (lo demás como v0.9.0) | Sí | T001, T002 | Idem; la viñeta de `jerarquia_normativa.md` no cambia |
| FR-003 (M1) | Sí | T001, T002 | Frase fuera del cuerpo y al final de `calculo_de_plazos.md` |
| FR-004 (tope y 10 882) | Sí | T002 | `TestSkillsDelRepositorio`, `TestVerticalDeEnsayo` (sin cambiar) y FR-005; ver A1 |
| FR-005 (cálculo desde `-sin-condicion`) | Sí | T002 | Campo `desde`; sale el subtest «hasta H41.1» |
| FR-006 (variantes sin cambiar) | Sí | T002 | Mutantes sobre las huellas de control y las tres formas |
| FR-007 (remisiones) | Sí | T003 | `fallosAdmitidosDelProducto` con una entrada |
| FR-008 (sondeo fuera del run) | Parcial por diseño | T002 | El arnés no cambia salvo `origenesDeLaVariante` (I1); el sondeo lo lanza la persona (US4, sin tareas) |
| FR-009 (guía) | Sí | T004 | Dos frases «hasta H41.1» de `docs/SKILLS.md` (líneas 142 y 160) |
| FR-010 (CHANGELOG) | Sí | T005 | Entrada tras la de H41 |
| FR-011 (evals en verde en el cierre) | No, por diseño | — | Lo mide el job del cierre del workflow; las filas `evals:` están en «Controles de umbral» |
| FR-012 (nada más cambia) | Sí | Todas | Lo guarda el guardián de diff y las rutas declaradas por tarea |
| FR-013 (guion de H41) | Sí | T002 | Líneas 11-13, 17-20 y 48-53; las he comprobado y coinciden con lo que describe la tarea |
| SC-001, SC-002, SC-005 | Sí | T001-T003 | Con sus controles de umbral |
| SC-003 (aceptación por sondeo) | No, por diseño | — | La mide la persona (ADR 0032); la propuesta no se fusiona sin su tabla |
| SC-004 | No, por diseño | — | Job de evals del cierre |

## Constitución

Sin conflictos. III (tests primero, offline): T001 escribe la suite congelada y ninguna tarea usa red ni sesiones con modelo.
VIII (skills primero): el hito cambia una skill medible con evals y su aceptación es el sondeo. Capa 3 (humano): el cambio de
un fichero existente de `testdata/` (`h41-skills-instaladas.txtar`) queda marcado `[datos]` y va al informe final. No hay
dependencias, órdenes ni banderas nuevas.

## Tareas sin requisito

Ninguna. T001 (aceptación), T002 (FR-001 a FR-006, FR-008, FR-012, FR-013), T003 (FR-007), T004 (FR-009) y T005 (FR-010) se
corresponden con algún requisito.

## Métricas

- Requisitos funcionales: 13 (más 5 criterios de éxito, de los que SC-003 y SC-004 se miden fuera del run)
- Tareas: 5
- Cobertura (FR con al menos una tarea): 12 de 13 = 92 %; el que falta, FR-011, es del job del cierre por diseño
- Ambigüedades: 1 (A1, LOW)
- Duplicaciones: 0
- Incidencias CRÍTICAS: 0 · ALTAS: 0 · MEDIAS: 2 · BAJAS: 3

## Próximos pasos

Sin incidencias críticas ni altas: se puede pasar a `implement`. Las dos medias (I1, C1) son decisiones ya registradas en
`gates/supuestos.md` y en los pendientes del plan; no piden cambiar los artefactos. Al implementar, comprobar la cifra de 10 882
(A1) con quickstart §4 antes de dar por buenos los mensajes de los mutantes.
