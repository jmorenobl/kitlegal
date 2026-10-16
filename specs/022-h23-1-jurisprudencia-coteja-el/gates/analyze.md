# Specification Analysis Report — H23.1

Feature: `specs/022-h23-1-jurisprudencia-coteja-el/` · Artefactos: `spec.md` (316 líneas), `plan.md` (396), `tasks.md` (349), con `research.md`, `data-model.md`, `contracts/` y `quickstart.md` como apoyo · Constitución 2.13.0.

**Cómo se hizo.** Lectura entera de `spec.md` y `tasks.md`; de `plan.md`, «Summary», «Aceptación e2e», «Controles mecánicos», «Controles de umbral», «Decisiones», «Trazabilidad» y «Orden de implementación»; de `quickstart.md` §4. Comprobado contra el árbol con `grep`/`jq`/`awk` (solo lectura). **No se ejecutó ningún test, `make ci` ni `precheck.sh`**: lo que sigue es coherencia entre artefactos y contra el código actual, no una medida de que el código por escribir pase.

## Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| I1 | Inconsistencia (declarada) | LOW | `.github/workflows/evals.yml:157`; spec.md «Fuera de alcance» | El comentario dice que cada una de las dos skills con juez lleva «trece umbrales, nueve que deciden». Tras el hito es falso para `jurisprudencia` (quince, once). Verificado: la línea existe y FR-052 impide tocarla. | Ninguna dentro del hito. Ya está como pendiente para la persona (spec «Fuera de alcance»; tasks, «Estrategia»). El informe final debe mantenerlo. |
| U1 | Subespecificación (declarada) | LOW | spec.md SC-010 y «Fuera de alcance»; FR-033 | El «0 de 30 sin cotejar» de la Aceptación no se lee en la salida de `make evals-sondeo` (mide solo el modo orden, no dice qué sesiones cotejaron). El hito lo deja como medida humana sin control en el run. | Ninguna dentro del hito (Clarification 2026-10-16, conservadora). Pendiente para la persona en el informe final. |
| C1 | Cobertura / tamaño de tarea | LOW | tasks.md T001 | T001 concentra código (`informe.go`, `umbrales.go`, `doc.go`) y los tests de cinco ficheros en una sola tarea. Está justificado y declarado (la forma de `umbrales` cambia y los tests que la esperan van con ella; plan «Orden», paso 1), pero es la tarea con más riesgo de agotar un intento. | Sin cambio. Si el run la parte, que sea por la frontera código/tests solo si `make ci` queda en verde en cada mitad (no lo queda: ver T001, «Por qué el umbral y todos sus tests son una sola tarea»). |
| C2 | Cobertura | LOW | tasks.md T005 (`cierre.md`) | T005 y su `cierre.md` no están entre los cuatro pasos del plan; tasks.md lo declara y lo justifica (punto 9 de la DoD, cobertura, sin quien lo mida en `make ci`). Rutas declaradas: `cierre.md`, `informe_test.go`, `umbrales_test.go` (estos dos solo si falta cobertura). | Sin cambio. |
| A1 | Ambigüedad residual | LOW | spec.md FR-003 («El texto exacto lo fija el plan») | El texto exacto del motivo y de la `descripcion` queda a `contracts/umbral-documento-sin-cotejar.md` §2–§3, no al spec. Es deliberado (spec «Assumptions», último punto) y los tests lo escriben a mano. | Sin cambio. |

Sin hallazgos CRITICAL, HIGH ni MEDIUM.

## Comprobaciones contra el árbol (sin ejecutar tests)

- Existen, con el nombre y fichero que citan plan y tareas: `TestUmbralesDelInforme` (`umbrales_test.go`), `TestUmbralesConJuezYConSentencias` y `TestUmbralConjunto` (`umbrales_test.go`), `TestUmbralesDeJurisprudencia`, `TestInformeFinalDeLosUmbrales` (subtest `como-en-el-cierre-de-h26`, `informe_test.go:1840`), `TestInformeEnDosModos`, `TestInforme` (`informe_test.go:267`), `TestInformeConElJuez`, `TestEvalsDelRepositorio` (`conjunto_test.go`), `TestJuzgar` (`juzgar_test.go:293`), `TestJuzgarLasLlamadas`, `TestJuzgarSentencias`, `TestCitaSinDocumento`, `TestCopiasDelJuez`, `TestMedidaVersionada`, `TestDefinicionDelJob`, `TestSalidaDelSondeo`, y `TestSkillsDelRepositorio` y `TestOrdenesDeLasSkillsEmpotradas` (`internal/app/skills_test.go`, en el objetivo `skills-check` del `Makefile`). El subtest `el-elemento-del-contrato` existe (`informe_test.go:5102`).
- Funciones de producción citadas: `satisface` (`juzgar.go:968`), `invocacionesDe` (`juzgar.go:493`, método de `*ResultadoDeEval`, devuelve `([]invocacionJuzgada, bool)`), `medirLasRespuestas` (`informe.go:1324`), `medirLasCitasSinDocumento` (`informe.go:1354`), `respuestasQueSeJuzgan` (`informe.go:1605`), `nombradasEn` (`umbrales.go:561`), `sinDocumentoDelUmbral` (`umbrales.go:577`).
- **Criterio de cotejo (FR-002) correcto sobre el código**: `satisface` exige `consultoConExito` (consulta, código 0, no `--dry-run`/`--describe`) y el mismo applet; para `cotejar` usa `pideLaCita`, que con `comando.ROJ == ""` y `ConTexto == false` solo mira el verbo. Un comando esperado `cita cotejar` sin ROJ cubre, pues, «cualquier argumento», como pide el spec. `formaDelComando` reparte `cotejar` en `formaCotejar` sin condición sobre norma o ROJ.
- Las cuatro evals que piden el cotejo son 04, 06, 08 y 09 (`grep 'verbo: cotejar' evals/`). 4 evals × 3 repeticiones = 12 sesiones por modo (coincide con el `total` del spec y del plan).
- `skills/jurisprudencia/SKILL.md`: 197 líneas; el límite de 300 lo comprueba `internal/skills/sincronia_test.go` (`defectos: "SKILL.md tiene 300 líneas (máximo 299)"`).
- Último informe versionado (`specs/021-h26-el-umbral-de/gates/evals/jurisprudencia.json`): 13 umbrales, 9 que deciden, en el orden que el plan toma como base; los nombres de las once filas `evals:` de «Controles de umbral» coinciden con los nombres reales (`sin_activar`, `cita_sin_documento`, `afirma_lo_no_leido:<modelo>`, `medida_del_juez:afirma_lo_no_leido:defectos_sin_marcar` y `…correctos_marcados`, `duracion_del_juez:<modo>`) más los dos nuevos.
- La tabla `casosDeLosUmbralesDeJurisprudencia` tiene hoy diez casos (`ninguna-cuenta` y nueve más): coincide con «diez casos de antes» y con «los catorce casos de la tabla» para el mutante 2 (10 + 4 nuevos).
- Frases de los documentos que T004 manda cambiar, presentes hoy: `CONTRIBUTING.md:1145` («otros trece»), `:1172` («son también trece, y deciden nueve»), `:1187` («los cuatro del modo orden…»), `:1190` («cuatro de los trece»); `docs/JURISPRUDENCIA.md:227` («el job decide con `cita_sin_documento`»); `CHANGELOG.md:13` (`## [Unreleased]`). `docs/WORKFLOW.md:139` solo habla de `boe-legislacion`, como dice FR-044. La enumeración de `CLAUDE.md` que FR-043 menciona está en el fichero.
- Mutantes de `quickstart.md` §4: `sed -i.orig` vale en BSD y GNU; lleva guarda «NO SE HA APLICADO» por si el receptor de `sinCotejar` no coincide con el patrón.

## Cobertura de requisitos

| Requisito | ¿Tarea? | Tareas | Notas |
|---|---|---|---|
| FR-001 a FR-007 (umbral) | Sí | T001; T002 (mutantes) | `piden-el-cotejo` en `TestEvalsDelRepositorio` cubre FR-001/FR-007 |
| FR-020 a FR-024 (skill) | Sí | T003 | Efecto lo mide el job de cierre (SC-001); forma, `make skills-check` |
| FR-030 a FR-034 (lo que sigue igual) | Sí (restricción y tests sin cambiar) | T001; T005 lo constata | `umbralesDeJurisprudenciaComoEnH26` mantiene el caso H26 |
| FR-040 a FR-045, FR-065 (documentos) | Sí | T004; T005 | `WORKFLOW.md` se lee y no cambia |
| FR-050 a FR-052 | Sí (restricciones del run) | batería de verificación; T005 | Rigen también las reparaciones del cierre |
| FR-060, FR-061, FR-063 | Sí | todas; plan «Controles de umbral»; T001 | Once filas `evals:` + filas `ci:` |
| FR-062, FR-064 | Sí | T001; T002 | Casos con sesión cotejando y otra no en el mismo modo (distingue «cada sesión») |
| FR-066 | Workflow | — | Lo hace el cierre, no una tarea |
| SC-001, SC-009 | Workflow | T001 construye los dos umbrales, T003 la skill | Se miden en el job de cierre y en el informe final |
| SC-002 a SC-006 | Sí | T001; T002 | |
| SC-007 | Sí | T003 | |
| SC-008 | Revisión final | T004 | Por diseño sin test que ate el texto (FR-045) |
| SC-010 | Persona, tras el run | — | Sin control en el run, declarado |

## Constitución — alineación

Sin conflictos. Comprobado: principio VIII (corrige una skill y protege su promesa con un control, sin herramienta nueva), «Umbrales que se cumplen» (ADR 0029: el umbral decide con 0, tiene fila en «Controles de umbral», test que lo ve fallar y mutante por control), ADR 0038 («cero donde hay forma»; el cotejo es un hecho de la sesión medido sin modelo y sin juez), ADR 0032 (ninguna tarea abre sesión con modelo, usa credenciales ni lanza `make evals`/sondeo), ADR 0036 (nada que consulte el CENDOJ), «Criterio de decisión autónoma» (lo no especificado, como cambiar el sondeo, se registra como supuesto y no se implementa). Ningún ADR nuevo, justificado en el plan.

## Tareas sin requisito

Ninguna. T005 (cierre) se mapea a FR-032 a FR-034, FR-045, FR-050 a FR-052, FR-060, FR-061 y SC-009.

## Métricas

- Requisitos funcionales: 33 (FR-001–007, 020–024, 030–034, 040–045, 050–052, 060–066); criterios de éxito: 10 (SC-001–010).
- Tareas: 5 (T001–T005); ninguna `[P]`, `[aceptacion]` ni `[datos]` (justificado en plan, «Aceptación e2e: no aplica» y «Datos externos»).
- Cobertura de requisitos con trabajo construible: 100 % (los que no son tarea —FR-066, SC-001, SC-009, SC-010— tienen su responsable declarado: workflow o persona).
- Ambigüedades: 1 (A1, baja y deliberada). Duplicaciones: 0. Hallazgos CRITICAL: 0. HIGH: 0. MEDIUM: 0. LOW: 5.

## Qué no se ha podido medir

- Si el código de T001 y sus tests pasan `make ci`, `golangci-lint` (`dupl`, `gocyclo`, `paralleltest`) o `-race`: no se ejecutó nada. Las cifras de `research.md` (tablas V y M) salen de un prototipo que ya no existe.
- Si `skills/jurisprudencia/SKILL.md` con la redacción de T003 hace que el job de cierre dé 0 de 12 en los dos modos (SC-001): lo mide el job; SC-010, una persona.
- `precheck.sh tasks` no se ejecutó en esta sesión; tasks.md dice que sus comprobaciones se reprodujeron con `grep`, y esta lectura no las ha repetido una a una.

## Siguientes acciones

Sin CRITICAL ni HIGH: se puede pasar a implementar. Ningún cambio a los artefactos es necesario; los dos pendientes de I1 y U1 son de la persona y el informe final debe conservarlos.

Hooks de extensión: `before_analyze` y `after_analyze` solo registran `speckit.git.commit` como opcional; en modo desatendido no se ejecuta (el commit lo hace el workflow).
