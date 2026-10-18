# Specification Analysis Report — H9 (`023-h9-internal-core-plazos`)

Análisis de `spec.md`, `plan.md` y `tasks.md`, con `research.md`, `contracts/` y la constitución como contexto. Solo lectura, salvo este fichero. **No se ejecutó ningún test ni `make ci`** y no se corrió `precheck.sh`. Lo que se comprobó contra el árbol se dice en cada fila; el resto (cifras de `research.md` salidas de un prototipo que ya no existe, lo que dirá lo grabado del BOE y del BOCM) no se pudo medir en esta sesión. De la constitución se comprobó su versión (2.13.0) y los puntos 8, 9, 11, 12 y 13 de la Definition of Done de `docs/ROADMAP.md`. No se releyó entera.

## Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| F1 | Inconsistencia | MEDIUM | spec.md FR-058 · tasks.md T011 | FR-058 prohíbe cambiar «los tests que leen su concurrencia» (comprobable por el diff). T011 edita `internal/evals/definicion_test.go` (`TestDefinicionDelJob`, que lee la definición del job y su concurrencia) para cambiar el literal `peorCasoDeLegalCore` y las líneas de dos subtests. FR-050 y la cifra 14 629 s piden ese cambio, así que los dos requisitos se leen en conflicto. | Aclarar en T011, y en el informe, que FR-058 se refiere a los valores de concurrencia y a los topes, no al literal del peor caso. T011 ya dice «ningún valor de concurrencia, ningún tope». Así un juez que lea el diff literalmente no lo toma por violación. |
| C1 | Cobertura / rutas | MEDIUM | tasks.md T004, T005 | T005 puede tener que ajustar el cómputo si lo grabado contradice al plan (art. 182.1 LOPJ sin el 24 y el 31 de diciembre, S4). Declara `computo.go`, pero los tests que fijan esas reglas están en `calcular_test.go` y `propiedades_test.go`, que T004 declara y T005 no. El guardián de diff solo admite el `x_test.go` de cada `x.go` declarado. | Añadir `internal/core/plazos/calcular.go` y `internal/core/plazos/propiedades_test.go` a las Rutas de T005, o cubrir el caso con una frase explícita. |
| C2 | Underspecification | LOW | tasks.md T016 | El punto (2) dice que, si la cobertura de `internal/core/plazos` queda bajo el 85 %, se añadan tests «añadiendo su ruta a esta línea». La tarea solo declara `cierre.md`, y las rutas se congelan por intento. T004 ya exige ≥ 85 % con `go test -cover`, pero T005 y T010 tocan después `computo.go`. | Declarar de antemano un fichero de test concreto para ese caso (p. ej. `internal/core/plazos/cobertura_test.go`) o quitar la cláusula y que el cierre solo informe. |
| I1 | Inconsistencia | LOW | spec.md Assumptions («Las listas literales…») y FR-085 · plan.md «Tests y guiones existentes que cambian» · research V11 | La suposición del spec dice que `internal/empaquetado/ejecutar_test.go` y `snapshot_piezas_test.go` nombran las herramientas y «cambian» con `plazos_calcular`. Plan, research V11 y tasks dicen que no cambian. Comprobado: el comentario de `herramientasDeHoy` en `ejecutar_test.go` dice «El test pide que estén, no que sean las únicas», y T009 y la batería excluyen `internal/empaquetado`. | Mantener lo del plan y las tareas. La suposición del spec queda superada por research V11. No hay que editar el spec (ya está fijado), pero un juez debe leerla así. |
| I2 | Ambigüedad | LOW | tasks.md T011 («⌈49 / 2⌉ + ⌈48 / 2⌉ + ⌈6 / 2⌉») · T015 («48, 48 y 6 más la prueba de red») · plan.md l. 77 | Las sesiones de `legal-core` se dan como 49/48/6 en el literal del peor caso y como «48, 48 y 6, más la prueba de red» en los documentos. Es compatible (el 49 sería el 48 más la prueba de red), pero ningún artefacto lo dice. T015 pide que cada cifra salga del árbol. | Que T015 saque las cifras del plan del job real (`TestDefinicionDelJob` o el plan de `internal/evals`) y no copie ninguna del plan. |
| U1 | Underspecification / dependencia de datos | LOW | tasks.md T007, T010 · spec.md FR-011, FR-025 | `TestFestivoDeUnaParte` (T010) falla si lo empotrado no trae ningún festivo autonómico limitado a una parte. Si las tres resoluciones grabadas no lo traen, T010 queda bloqueada. FR-011 ya está cubierto con calendarios sintéticos en T003 y T004; lo que depende de los datos es solo la prueba con lo real. | Aceptable: T007 prevé anotarlo en `gates/supuestos.md`. Conviene que el ejecutor sepa que el fallo no es un defecto del código. |
| U2 | Cobertura sin control mecánico | LOW | spec.md FR-040, FR-041 (disposición), FR-043, FR-044, US1.3 | Se comprueban solo por lectura de los jueces de la revisión final (la skill pregunta el municipio que falta y nunca usa el de quien usa el kit, el acto presunto, el art. 24.6 LTAIBG, el plazo del silencio). El spec lo declara y FR-050 impide añadir evals. | Sin acción. Queda como riesgo aceptado y declarado. |
| U3 | Alcance inferido | LOW | tasks.md T006 | T006 añade `Registro.MunicipiosDe`, `Registro.Comunidades` y `FormaCompacta` a `internal/core/territorio` (código de H6). Ningún FR lo nombra; se justifica por FR-022 (cruce por código INE) y por tener una sola definición del pliegue. No cambia `territorio resolver` ni su `--describe`. | Sin acción. Vigilar en la revisión final que no cambie el comportamiento de `resolver` ni su golden. |

No hay hallazgos CRITICAL ni HIGH. Resultado de las otras pasadas:

- **Duplicación**: ninguna que merezca fila. El spec repite a propósito reglas entre FR (p. ej. FR-006/FR-010 y el mutante de FR-089).
- **Ambigüedad / marcadores**: sin `TODO`, `TKTK`, `???` ni `<placeholder>` en `spec.md`, `plan.md` ni `tasks.md`. Los términos vagos que hay tienen cifra (0 sesiones, ≥ 85 %, < 300 líneas, 179 de 179).
- **Terminología**: coherente en los tres artefactos (`plazo_sin_fecha`, `cobertura_sin_declarar`, `sin_activar`, `plazo`, `no_cubierto`, `calcular`). El spec deja los nombres al plan y las tareas usan los del plan. El orden de `umbrales` de `legal-core` (`sin_activar`, `plazo_sin_fecha`, `cobertura_sin_declarar` por modo) coincide con `contracts/evals-legal-core.md`.
- **Orden de tareas**: la tabla de dependencias no tiene ninguna hacia delante. T005 depende de T002 y T003. T007 y T008 dependen de T006. T009 depende de T004, T005, T007 y T008. T010 depende de T008 y T009. T012 depende de T011. T013 depende de T012.

## Comprobado contra el repositorio (solo lectura)

- Existen con el nombre que usan las tareas: `TestSkillsDelRepositorio` (`internal/app/skills_test.go`), `TestHerramientasDelServidor` (`herramientas_test.go`), `TestArquitectura` (`internal/arch_test.go`), `TestDefinicionDelJob`, `TestConjuntoDeEvals` y `TestEvalsDelRepositorio` (`internal/evals`).
- `evals/legal-core/` tiene hoy cuatro evals (01 a 04). Los tests que esperan `umbrales` de `legal-core` vacío están en `informe_test.go` y `umbrales_test.go`, que son los `x_test.go` de los ficheros que declara T012. Ningún otro test cuenta las evals de `legal-core`, comprobado por `grep`.
- Evidencia de datos: `data/territorio/municipios.yaml` da Brunete `28026`, Majadahonda `28080` y «Vellón, El» `28168`. Brunete es «no comunicado» en `BOCM-20251212-34` y tiene línea en `BOCM-20251229-19` (15 de mayo y 14 de septiembre). Majadahonda es 14 de septiembre en `BOCM-20241213-23` y 15 en `BOCM-20250203-16`. Las correcciones de 2026 no mencionan Leganés, así que la fuente `BOCM-20251212-34` de su 9 de octubre en `plazos-calcular.txtar` se sostiene con lo que ya hay en `evidencias/adr-0039/`.
- Fechas de los guiones de T001, calculadas a mano en esta sesión: 9 de septiembre más un mes da el 9 de octubre de 2026 (viernes, fiesta local de Leganés). El 10 y el 11 son fin de semana y el 12 es nacional, así que vence el 13. 15 de junio de 2026 más dos meses sin agosto da el 15 de septiembre (martes). 27 de octubre más dos meses da el domingo 27 de diciembre y se prorroga al lunes 28. 30 de diciembre de 2024 más tres hábiles da el 3 de enero de 2025. Coinciden con T001.
- DoD de `docs/ROADMAP.md`: el punto 9 (≥ 85 % en `internal/core/**`) lo cubren T004 y T016. El 11 y el 13 aplican y están en la tabla de tareas. El 12 no aplica (FR-013). Para el 8, FR-028 y las filas ya revisadas de `docs/SOURCES.md` lo dejan sin cambios; la decisión está en el spec y el plan.

## Resumen de cobertura

| Grupo | Tiene tarea | Tareas | Notas |
|---|---|---|---|
| FR-001 a FR-013 (cómputo y applet) | Sí | T001, T003, T004, T005, T009, T010 | |
| FR-020 a FR-028 (festivos) | Sí | T002, T003, T006, T007, T008, T009 | FR-028 también por T016 |
| FR-030 a FR-032 (tipos y cobertura por materia) | Sí | T002, T003, T005, T013 | |
| FR-040 a FR-048 (skill) | Sí | T004, T011, T013 | FR-040, FR-043 y FR-044: control solo por jueces (U2) |
| FR-050 a FR-058 (evals y umbrales) | Sí | T002, T005, T011, T012 | FR-058: ver F1 |
| FR-060 a FR-063 (documentos) | Sí | T015, T016 | |
| FR-070 a FR-073 (lo que no cambia) | Sí | batería, T012, T016 | revisión final |
| FR-080 a FR-090 (controles) | Sí | todas, T004, T007, T008, T010, T012, T014, T015 | FR-086 es del plan y T012 publica sus umbrales |
| FR-091 | Fuera de tareas por diseño | informe final del workflow; T016 lo constata | |
| SC-001, SC-002 | Fuera de tareas por diseño | los mide el job de cierre y el informe final | |
| SC-003 | Ninguna (humano, tras el run) | — | declarado en el spec |
| SC-004 a SC-013 | Sí | T004, T007, T008, T009, T010, T012, T013, T014, T015 | SC-013 por la revisión final |

Tareas sin requisito: ninguna. T001, T014, T015 y T016 son de aceptación, de verificación o de cierre y citan sus FR y SC.

## Alineación con la constitución

Sin conflictos. Genericidad territorial (principio IX): ningún municipio con caso especial, y la batería lo recuerda; T016 lo comprueba. Skills sin código y `SKILL.md` < 300 líneas: T013. Sin red en los tests, sin sesiones con modelo y umbrales con 0 donde hay forma (ADR 0038): T012 y la batería. Los datos externos los graba `grabar_datos` desde los manifiestos de T002, sin red ni memoria en las tareas. Ninguna tarea cambia `.github/workflows/`, y los umbrales nuevos son `decide: true` con 0.

## Tareas sin mapa

Ninguna.

## Métricas

- Requisitos funcionales: 63 (FR-001 a FR-091). Criterios de éxito: 13.
- Tareas: 16 (T001 a T016).
- Cobertura de FR con al menos una tarea o control declarado: 63 de 63. Tres se cumplen fuera de las tareas por diseño (FR-091, SC-001/SC-002 con el workflow) y uno es humano (SC-003).
- Ambigüedades: 1 (I2). Duplicaciones: 0. Inconsistencias: 2 (F1, I1). Huecos de rutas o cobertura: 3 (C1, C2, U1).
- Hallazgos CRITICAL: 0. HIGH: 0. MEDIUM: 2. LOW: 6.

## Próximas acciones

- No hay bloqueos para `implement`. Los dos MEDIUM (F1, C1) se resuelven en `tasks.md` con una línea cada uno: aclarar el alcance de FR-058 en T011 y ampliar las Rutas de T005.
- Orden sugerido si se retoca `tasks.md`: C1 (rutas de T005), F1 (T011), C2 (T016) e I2 (T015).
- Fuera de este análisis, el workflow decide si repite el gate de tareas tras esos retoques. No se aplicó ningún cambio.

## Hooks de extensión

`.specify/extensions.yml` tiene `before_analyze` y `after_analyze` con `speckit.git.commit`, ambos opcionales. En modo desatendido no se ejecutan: el commit lo hace el workflow.
