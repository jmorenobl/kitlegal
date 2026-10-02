# Specification Analysis Report — H22 (016-h22-instalar-sin-terminal)

Análisis de solo lectura de `spec.md`, `plan.md` y `tasks.md` contra la constitución (2.10.0), `docs/ROADMAP.md` (sección H22) y el árbol de `main` (`9e9a993`). Sin modo de plataforma, sin red, sin sesiones con modelo.

**Qué se ejecutó y vio terminar en esta sesión**: lecturas de los tres artefactos, de `gates/supuestos.md`, de `contracts/paso.md` y `contracts/release.md` (partes), de `docs/ROADMAP.md` H22 y de la constitución; `grep` sobre `internal/app`, `internal/arch_test.go`, `snapshot_test.go`, `.github/workflows/release.yml`, `evals.yml`, `Makefile`, `.gitignore`; y el extractor de rutas de `scripts/workflow/tarea.sh` (línea 141) aplicado a las siete líneas de tarea, más la lectura del comparador de `scripts/workflow/guardian-diff.sh` (líneas 75-100). Un fichero temporal en la raíz del repositorio que usé para lanzar el extractor se borró en la misma sesión (`git status` solo muestra `tasks.md` sin seguir, como al empezar).

**Qué no se pudo medir**: no se ejecutó `make ci`, `make release`, `make snapshot-check`, `goreleaser`, `claude plugin validate` ni ninguna orden de `humo` o `catalogo`: lo que dicen research V1-V32 y S1-S11 sobre ellos no se ha reverificado aquí y se cita como lo que el plan afirma, no como evidencia propia.

## Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| C1 | Cobertura / Constitución (criterio 2, ADR 0029) | HIGH | docs/ROADMAP.md H22 «Controles» (2.º punto); spec.md Clarifications Q2, FR-017, FR-061; plan.md «Datos externos» | El control literal del hito —el manifiesto cumple el esquema `0.3` «versionado en `testdata/`»— no se construye: se sustituye por una comprobación propia (campos de FR-012, `version`, `tools`). La constitución dice que dejar fuera el control de un umbral del hito «es un umbral que nadie cumple». La decisión es deliberada, está registrada (Clarifications Q2, `gates/supuestos.md`, «Fuera de alcance») y su causa es real (el esquema no puede llegar a `testdata/` en el run: sin red y sin fila en `SOURCES.md`, ADR 0018). `docs/ROADMAP.md` no se enmienda. | No cambiar artefactos. El informe final debe nombrarlo entre los supuestos de impacto `alcance`, y la persona decidir si versiona el esquema o enmienda el control del roadmap. No es CRITICAL porque no contradice un MUST: está sustituido por un control equivalente en lo que el run puede medir y lo dice. |
| C2 | Inconsistencia | MEDIUM | spec.md FR-030, «Key Entities» («La plantilla del catálogo: el fichero de este repositorio»); plan.md «Decisiones» y `supuestos.md` | El spec pide un fichero de plantilla en el repositorio; el plan lo implementa como código (`internal/empaquetado/catalogo.go`) sin fichero JSON versionado. El plan lo registra como supuesto (impacto `comportamiento`), pero el spec no se actualiza y la entidad «plantilla» del spec no tiene objeto correspondiente. | Dejarlo como está para el run (se recoge en el informe final) o, al relanzar, ajustar FR-030 a «el código que genera el catálogo». |
| C3 | Riesgo no medido | MEDIUM | research S3, S8; spec Assumptions («`claude plugin validate`… no una medida»); tasks T004 | Que `claude plugin validate` acepte un catálogo con entrada `archive` y el plugin, y que `npm install -g @anthropic-ai/claude-code@2.1.284` funcione en `ubuntu-latest`, solo se verá en el trabajo `snapshot` de la propuesta de cambio, después de la revisión final. No hay tarea que lo ejerza ni alternativa prevista si falla. | Es lo que el plan declara (el cierre lo cuenta). Si falla, entra por el bucle de cierre (ADR 0030); sin acción previa. |
| C4 | Riesgo no medido | MEDIUM | research S1, S2, S5, S10, S11; tasks T005 | Los pasos nuevos de `humo`, el trabajo `catalogo` y la atestación de dos ficheros más solo se ejecutan con una etiqueta que empuja una persona. `TestConfiguracionDeLaRelease` fija su definición y `bash -n` su sintaxis, no su comportamiento. El plan lo reconoce y T005 obliga al informe a decirlo. | Sin acción; asegurar que el informe final lo repita (T005 ya lo exige). |
| I1 | Inconsistencia menor | LOW | spec.md FR-041 («Esa separación con órdenes de shell no se ha medido») frente a plan.md «Technical Context» y research S1 («medidos con bash 3.2», en macOS) | El spec dice que la separación de arquitecturas con `od`/`head`/`tail` no se ha medido; el plan la midió en macOS con el prototipo, y lo que queda sin medir es Linux (S1). | Ninguna: el spec es anterior a la medida; la frase queda obsoleta pero no cambia ningún requisito. |
| I2 | Ambigüedad de redacción | LOW | tasks.md T004 («por su función, importando el paquete como `paso`») frente a contracts/paso.md §6 («Nada más se exporta» que `Ejecutar` y cuatro textos) | T004 habla de escribir el catálogo «por su función», pero el paquete solo exporta `Ejecutar`: la única vía es `paso.Ejecutar([]string{"catalogo", …}, …)`, que es lo que dice `contracts/release.md` §4 («con lo que da `empaquetar catalogo`»). Se resuelve sin tocar nada, pero la frase puede hacer que el ejecutor busque una función exportada que no existe o la exporte. | Si se edita tasks.md: «llamando a `paso.Ejecutar` con la orden `catalogo`». |
| I3 | Terminología | LOW | spec.md «Vocabulario» («el paso» = el programa) y uso de «paso» para los pasos de `humo`, del workflow y de `ci.yml` | «Paso» designa el programa, los pasos de un trabajo de GitHub Actions y los pasos del workflow `hito`. El spec lo define, pero en FR-040/FR-041/FR-031 («el paso que escribe el catálogo») se mezclan los dos primeros. | Sin acción: el contexto lo desambigua. |
| U1 | Infraespecificación | LOW | spec.md US3 escenario 3, FR-005; tasks T002/T003 | FR-005 dice que la release o el snapshot «fallan» si el paso sale con código ≠ 0. Los tests cubren el código de salida del paso (`TestPiezasSinEntrada`, `TestEjecutar`); que goreleaser aborte con el gancho lo respalda research V4 (medido con el prototipo) pero ninguna tarea lo ejerce. | Sin acción: depende del comportamiento de goreleaser v2.18.1 ya medido. |
| T1 | Tareas | LOW | tasks.md T002 | Tarea grande (paquete entero, `main`, regla R1, siete tests, dos listas de arquitectura). Declarada y justificada en la propia cabecera (sin el `main`, `unused` dejaría `make ci` en rojo). | Sin acción. |
| T2 | Tareas / guardián | LOW | tasks.md T002, T003, T004, T005, T006 (líneas de tarea) | El extractor de rutas del workflow declara también lo que cada línea cita como contexto: `archive/zip`, `image/png`, `debug/macho`, `contracts/paso.md`, `contracts/release.md`, `contracts/documentacion.md`, `plan.md`, `quickstart.md` y nombres sueltos (`install.sh`, `checksums.txt`, `evals.yml`, `SKILL.md`, `SOURCES.md`…). El comparador del guardián exige coincidencia exacta de ruta o prefijo de directorio, así que ninguna de ellas casa con un fichero real: no amplía lo que una tarea puede tocar (en particular, `evals.yml`, `SKILL.md` y `SOURCES.md` quedan no declarados). Ninguna línea contiene `testdata/`, `schemas/` ni `[datos]`/`[plataforma]`/`[aceptacion]` (comprobado con el extractor y con `grep`). | Sin acción. |

## Tabla de cobertura

41 requisitos funcionales (FR-001–006, 010–017, 020–023, 030–032, 040–043, 050–053, 060–070, 080) y 12 criterios de éxito (SC-002 humano, SC-012 documental).

| Requisito | ¿Tarea? | Tareas | Notas |
|---|---|---|---|
| FR-001 | Sí | T002, T003 | `cmd/empaquetar`, `TestElBinarioNoEnlazaElPaso`, gancho |
| FR-002, FR-006 | Sí | T003 | `extra_files`, `universal_binaries`, `archives.ids` |
| FR-003 | Sí | T005 | nueve sujetos atestados |
| FR-004, FR-067 | Sí | T002 | `TestPiezasReproducibles` |
| FR-005 | Sí | T002, T003 | `TestPiezasSinEntrada`; el aborto de goreleaser solo por research V4 (U1) |
| FR-010–FR-013, FR-017 | Sí | T002, T003 | `TestPiezas` + subpruebas del snapshot; esquema oficial fuera (C1) |
| FR-014 | Sí | T001, T002, T003 | `HerramientasAnunciadas`; `servidor-de-la-extension` |
| FR-015, FR-016, FR-066 | Sí | T002, T003 | `TestDescripcionCorta`, `TestIcono` |
| FR-020–FR-022 | Sí | T002, T003 | `skills-del-plugin` |
| FR-023 | Sí | T004 | solo en la CI (C3) |
| FR-030 | Sí | T002, T004 | plantilla como código (C2) |
| FR-031, FR-032 | Sí | T002, T005 | primera ejecución real = primera release (C4); permiso del token, de la persona |
| FR-040–FR-043 | Sí | T005 | solo se fija la definición (C4) |
| FR-050–FR-053 | Sí | T006, T007 | verificación por revisión final |
| FR-060–FR-064 | Sí | T003 | `make snapshot-check` |
| FR-065 | Sí | T004 | `make plugin-check` en el trabajo `snapshot` |
| FR-068 | Sí | T003, T004, T005, T007 | `TestConfiguracionDeLaRelease` |
| FR-069 | Sí | T001–T003, T007 | `depguard` R1, cobertura |
| FR-070 | Sí | T002; tabla de «Controles de umbral» de tasks.md; T007 | seis umbrales con control en `make ci` |
| FR-080 | Sí | T001, T007 | nada de lo anunciado cambia |
| SC-001 | Sí | cierre del workflow; T007 | mide el trabajo `snapshot` |
| SC-002 | No (humano) | — | fuera del run por diseño |
| SC-003–SC-007, SC-009–SC-011 | Sí | T002–T005 | `make ci` y snapshot |
| SC-008 | Sí | T004 (control), cierre (medida) | |
| SC-012 | Sí | T006 | sin umbral numérico |

## Alineación con la constitución

- **Sin violaciones de MUST.** Principios I–IX y reglas R1–R7: tabla completa en plan.md; verificado en el árbol que `paquetesInternos` hoy tiene diez entradas (`internal/arch_test.go`, línea 869) y el plan lo lleva a once; que `app.RegistroDeProduccion` (`internal/app/registro.go`), `kitlegal.Skills()` (`skills.go`), `verbosAnunciados` y `NombresDeHerramientas` (`internal/app/herramientas.go`) y `TestEntregaDelHito` existen; que `TestSnapshot` tiene hoy cuatro subpruebas (las seis nuevas dan las diez de T003); que `release.yml` atesta hoy siete sujetos y `humo` tiene seis pasos (el plan pasa a nueve y doce); que `VERSION_DE_CLAUDE_CODE` es `2.1.284` en `evals.yml`; que `/dist/` está en `.gitignore`; y que `make snapshot-check` y `make release` existen como los describe el plan.
- Reglas del modo desatendido: ninguna tarea toca `testdata/`, `schemas/`, `data/` ni `evidencias/`; ninguna etiqueta de datos/plataforma/aceptación (precheck de `tasks` coherente con «Aceptación e2e: no aplica»); ninguna tarea ejecuta `claude`, `make plugin-check`, `make evals` ni `release.yml`.
- Criterio de decisión autónoma, punto 2: ver C1.
- ADR 0029: los umbrales de FR-070 tienen fila y control en `make ci` en plan.md; las cinco medidas que solo existen sobre el snapshot real tienen su control en `make snapshot-check` / `make plugin-check` y el plan lo declara.

## Tareas sin requisito

Ninguna. T007 (cierre de la Definition of Done) no construye un requisito pero cubre FR-053, FR-068–FR-070 y FR-080 como comprobación.

## Métricas

- Requisitos funcionales: 41; criterios de éxito: 12 (11 con tarea o cierre; SC-002 humano).
- Tareas: 7 (T001–T007), ninguna `[P]`, ninguna de datos ni de aceptación.
- Cobertura de requisitos con ≥ 1 tarea: 41/41 = 100 % (SC-002 excluido por humano).
- Ambigüedades: 2 (I2, I3).
- Duplicaciones: 0 (los pares FR-011/FR-062, FR-012–014/FR-061 son requisito y control, a propósito).
- Hallazgos CRITICAL: 0. HIGH: 1 (C1). MEDIUM: 3 (C2, C3, C4). LOW: 6 (I1, I2, I3, U1, T1, T2).

## Siguientes acciones

- Sin CRITICAL: puede pasar a `/speckit-implement`.
- Lo que el informe final debe hacer visible: C1 (esquema oficial no validado; propuesta de cambio o enmienda del roadmap, de la persona), C2 (plantilla como código), y que C3 y C4 son riesgos que solo se miden en el trabajo `snapshot` de la propuesta de cambio y en la primera release.
- Si se quiere pulir antes de implementar (opcional): I2, frase de T004; I1, frase de FR-041.
