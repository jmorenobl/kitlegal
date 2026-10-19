# Specification Analysis Report — H20 · `boe articulo --fecha` + `boe-legislacion` v0.2

Artefactos analizados: `spec.md` (368 líneas), `plan.md` (489), `tasks.md` (434), y como apoyo `contracts/` y `research.md` por grep. Sesión de un paso (ADR 0032): solo lectura de ficheros y `grep`; **no se ha ejecutado ningún test, `make ci` ni el binario**, y no se ha leído entera la constitución (se ha confiado en la «Constitution Check» del plan y en los gates anteriores de spec y plan). Lo comprobado en el árbol en esta sesión: que existen `TestClavesDeLosDatos` (`internal/source/boe/datos_test.go:129`), `TestLeerBloque`, `TestFallosNoSeGuardan`, `TestObservadoDeBoe`, `articuloDelBloque` (`internal/source/boe/articulo.go:234`), `maximoDeEvals = 21` (`internal/evals/conjunto.go:227`), `TestConjuntoDeEvals`, `TestEvalsDelRepositorio`, `TestDefinicionDelJob`, `TestUmbralesDeLasTresSkills`, `umbralesDeBoeLegislacionComoEnH26`, `fechasDeVigencia` (`redacciones.go:62`), `formaDeCita` (`citas.go:16`), `nombradasPorSuSesion` (`umbrales.go:776`), y las veintiuna evals de `evals/boe-legislacion/`.

Resultado: **ningún hallazgo crítico**. Un hallazgo de severidad media y tres bajos.

## Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| I1 | Inconsistencia | MEDIUM | spec.md FR-059 (L263) y FR-050 (L244) · tasks.md T010 (L219) | FR-059 prohíbe cambiar «los tests que leen su concurrencia» y se comprueba con el diff del hito; FR-050 y T010 exigen cambiar `internal/evals/definicion_test.go` (el literal `peorCasoDeBoeLegislacion`, hoy en L1874, y su caso en L1952), el fichero de `TestDefinicionDelJob`, que lee el flujo `evals.yml` (concurrencia, tope). Leído al pie de la letra, el diff de T010 incumple FR-059. | Entender FR-059 como «no cambia ningún valor de concurrencia, objetivo ni tope ni la lógica de esos tests», y T010 toca solo el literal y su caso (ya lo dice: «la concurrencia (5), el objetivo de 900 s y el tope de 352 minutos, iguales»). Que la revisión final lea el diff de `definicion_test.go` con esa regla, y no como infracción de FR-059. |
| I2 | Inconsistencia | LOW | spec.md SC-004 (L311) y FR-081 (L281-283) · plan.md L293 y L207 · tasks.md T005 | SC-004 y FR-081 fijan **seis** casos del golden (cinco del `a22`, uno del `a21`); el plan y T005 los cuentan como **siete** (`TestArticuloEnUnaFecha`, «los siete casos de contracts §2»), porque el séptimo es el `a1-30` de US1. La fila del plan cita SC-004 con «7 casos». | No afecta a la ejecución. Si se quiere la cifra exacta, que el plan diga «los seis de FR-081 más el de US1.1». |
| A1 | Ambigüedad | LOW | spec.md FR-002 (L189) · tasks.md T005 (L186) | FR-002 elige «la de mayor `fecha_vigencia` que no es posterior a la pedida»; T005 elige «la última, **en el orden del bloque**, cuya `fecha_vigencia` no es posterior». Coinciden solo si el BOE da las versiones en orden creciente de vigencia (lo es en los tres bloques grabados según research M1). Con dos versiones de la misma vigencia o desordenadas difieren, pero el plan declara ese estado sin vía real (proporcionalidad). | Dejar la redacción de T005 y asegurar que `TestLeerBloque` fije el orden creciente de `fecha_vigencia` de los tres bloques (T003 ya fija las fechas de las nueve versiones del `a22`); así un cambio de orden del BOE rompe un test y no la elección. |
| T1 | Tamaño de tarea | LOW | tasks.md T010 (L219) | T010 reúne dos evals YAML, reglas del conjunto, medida e informe, umbral, `doc.go`, tests nuevos de informe, cambios en `TestUmbralesDeLasTresSkills`, `TestConjuntoDeEvals`, `TestEvalsDelRepositorio`, `TestConsultasNecesarias` y el literal de `TestDefinicionDelJob`; ocho rutas declaradas más sus `_test.go`. Está justificada (cambian a la vez el conjunto, la caché, el peor caso y los umbrales) y su diff es acotado por rutas, pero es la tarea con más riesgo de agotar una sesión de un paso. | Sin cambio. Si el run la corta por tiempo, el orden natural de partirla es: evals + reglas del conjunto + peor caso (con el umbral aún sin medir no cabe: cambia el conjunto de umbrales en `TestUmbralesDeLasTresSkills`), por lo que la partición real sería evals y umbral juntos y la medida de `fechadas` por separado; no se recomienda partirla de antemano. |

## Cobertura de requisitos

Inventario: 58 FR (FR-001…FR-092, con los huecos de numeración del spec) y 12 SC = 70 claves. Todas aparecen en «Requisitos → tareas» (tasks.md L288-342) o en el «Controles de umbral» del plan.

| Clave | ¿Tarea? | Tareas | Notas |
|-------|---------|--------|-------|
| FR-001…FR-007 | Sí | T001, T004, T005, T006 | FR-004 solo en T005 (tests construidos en el test) |
| FR-010…FR-014 | Sí | T001, T003, T005, T006 | |
| FR-020…FR-022 | Sí | T001, T005 | |
| FR-030…FR-032 | Sí | T001, T002, T007 | |
| FR-040…FR-046 | Sí | T011 | sin umbral propio salvo FR-054; los lee la revisión final (declarado) |
| FR-047…FR-049 | Sí | T006, T010, T011 | |
| FR-050…FR-059 | Sí | T008, T009, T010, T014 | FR-059: ver I1 |
| FR-060…FR-063 | Sí | T013, T014 | |
| FR-070…FR-072 | Sí | T001, T006, T010, T014 y la batería | |
| FR-080…FR-091 | Sí | todas; T001, T005, T006, T007, T012 | FR-088: filas del plan; T010 publica el umbral |
| FR-092 | Sí (constatación) | T014 | es una nota del informe final, no trabajo |
| SC-001, SC-002 | Sí (indirecto) | T010, T011; el job y el informe final | los mide el workflow |
| SC-003 | No, por diseño | — | humano, fuera del run (declarado) |
| SC-004…SC-012 | Sí | T001, T002, T005…T013 | SC-012 por lectura de la revisión final |

Cobertura: 69 de 70 con al menos una tarea o control (98,6 %); la que falta es SC-003, que el spec declara humana y sin control en el run.

## Constitution Alignment Issues

Ninguno encontrado en los artefactos. Comprobado por lectura de tasks.md: tests en la misma tarea que el código (T002 a T011), excepciones declaradas (T001, T012 a T014), una sola `[aceptacion]` y la primera, etiqueta `[datos]` solo donde se toca `schemas/`, sin grabaciones nuevas, sin ADR (justificado), controles de umbral con fila en el plan y su mutante, ninguna tarea que mida en la plataforma, que use la red o abra sesiones con modelo. Límite: la constitución 2.13.0 no se ha vuelto a leer entera en esta sesión.

## Unmapped Tasks

Ninguna. Cada tarea cita FR/SC. T012 (mutantes), T013 (documentos) y T014 (cierre) se mapean a FR-090, FR-060 a FR-063 y FR-080/FR-092.

## Chequeos de consistencia entre artefactos (sin hallazgo)

- Cifras de umbrales: spec FR-056/SC-009 (15 en `boe-legislacion`, 11 deciden; 6/6 en `legal-core`; 15/11 en `jurisprudencia`) = plan «Controles de umbral» = T010 y su tabla; 9 que deciden hoy + 2 nuevos = 11.
- Cifras del job: veintitrés evals, peor caso de 18 977 s (plan L61 y L302, research S1, `contracts/documentos.md`, T010) ≤ tope de 352 min (21 120 s); 60 sesiones por modo en `sin_activar` y 120 respuestas en `afirma_lo_no_leido` (spec FR-058, SC-001; plan; T010).
- Orden de dependencias: la tabla de tasks.md no tiene ninguna dependencia hacia delante; T005 → T006 → T007 y T008 → T009 → T010 → T011 respetan «evals antes que la skill» (DoD §1.10).
- Rutas declaradas frente a los símbolos tocados: `TestClavesDeLosDatos` compara los tipos de cada verbo con `ElementsMatch`; T005 añade `ArticuloFechado` a `datos.go` sin ponerlo como `Salida` (lo hace T006, que declara `datos_test.go`), así que no rompe el test entre una tarea y otra. Los símbolos que T008 y T010 leen sin editar (`fechasDeVigencia` en `redacciones.go`, `formaDeCita` en `citas.go`) no necesitan declararse.
- Regla del conjunto «entre 10 y 21 evals, exactamente 10 positivas» (research V29): T010 aparta las evals con `redaccion_fechada` de `positivas` y sube `maximoDeEvals` a 23, como la eval sin binario ni servidor.

## Métricas

- Requisitos (FR + SC): 70 (58 FR, 12 SC)
- Tareas: 14 (T001 aceptación, T002 a T011 rebanadas completas, T012 a T014 pulido y cierre)
- Cobertura: 98,6 % (69/70; SC-003 humana por diseño)
- Ambigüedades: 1 (A1)
- Duplicaciones: 0
- Inconsistencias: 2 (I1 media, I2 baja)
- Hallazgos críticos: 0

## Siguientes acciones

- Sin críticos: se puede pasar a `implement`.
- I1: que la revisión final y los jueces lean el diff de `definicion_test.go` con la interpretación de FR-059 de arriba (solo el literal y su caso). No requiere editar spec, plan ni tareas.
- I2, A1 y T1: sin cambio necesario; A1 queda cubierto si `TestLeerBloque` fija el orden de las versiones de los tres bloques.

## Extension Hooks

El hook `before_analyze` y el `after_analyze` de `.specify/extensions.yml` (`speckit.git.commit`) son **opcionales**. No se han ejecutado en esta sesión de un paso: el workflow `hito` hace sus propios commits.
