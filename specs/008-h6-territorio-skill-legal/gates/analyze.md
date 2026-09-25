# Specification Analysis Report — H6 · `territorio` + skill `legal-core` v0

Artefactos: `spec.md`, `plan.md`, `tasks.md` (con `research.md`, `data-model.md`, `contracts/`, `quickstart.md`) y la constitución. Análisis de solo lectura.

## Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| C1 | Infraespecificación | MEDIA | tasks.md T002 | T002 escribe los 19 ficheros de comunidad (nombres, régimen, provincias) y la configuración del BOCM (código, nombre, URL) sin nombrar de dónde salen. La batería prohíbe escribir códigos, títulos o identificadores «de memoria» y T001/T003 dicen que los aporta una persona; T002 no lo dice ni indica detenerse si falta el material. | Declarar en T002 el origen (provincias y códigos desde `municipios.yaml`; régimen, nombre y URL del BOCM revisados por la persona en la pausa) y añadir la parada en `gates/tarea-T002.md`. |
| C2 | Riesgo de delimitación | MEDIA | tasks.md T016 | T016 concentra ~10 ficheros de tres áreas (datos de normas, lector de jerarquía, tabla de generadores con `frontmatter.go`/`sincronia.go`/`referencias.go`, `Makefile`, regeneración de `normas.md`). Es la tarea con más probabilidad de agotar el intento o de que el guardián rechace una ruta. | Mantenerla si se justifica su indivisibilidad; si no, separar «lector de jerarquía + `Makefile`» de «tabla de generadores + normas vertebrales». |
| C3 | Orden / verificación | BAJA | tasks.md T019→T021 | T019 crea `evals/legal-core/` antes de que exista `skills/legal-core/`. No consta que `TestEvalsDelRepositorio` ni `TestSkillsDelRepositorio` sean indiferentes a una carpeta de evals sin skill. | Confirmar en T019 que ningún control existente exige carpeta de skill por carpeta de evals; si lo exige, anotarlo y redelimitar. |
| C4 | Robustez | BAJA | tasks.md T012 | El guion e2e impone `cronometra` < 200 ms (SC-004) en `make ci`; en un runner cargado puede ser intermitente (ver historial de flaky en `httpx`). | Comprobar que la orden usa la holgura ya establecida para otros guiones; no rebajar el umbral. |
| C5 | Consistencia menor | BAJA | tasks.md «Batería» vs T002 | La batería dice que la relación del INE, el REL y las grabaciones las obtiene una persona en T001, T002, T003 y T015; T002 no descarga nada. | Cuadrar la frase con C1 (T002 aporta revisión humana, no descarga). |

## Tabla de cobertura

| Grupo de requisitos | ¿Tiene tarea? | Tareas | Notas |
|---|---|---|---|
| FR-001–FR-016 (applet, sobre, exit codes) | Sí | T006, T009, T010, T012 | Trazabilidad explícita |
| FR-020–FR-024 (cobertura, genericidad) | Sí | T003, T006, T007, T009, T011, T012, T019 | |
| FR-030–FR-035 (ids INE/DIR3, fuzz) | Sí | T004, T005 | |
| FR-040–FR-049 (datos congelados, DIR3, SOURCES) | Sí | T001–T003, T007, T014, T016, T023 | |
| FR-050–FR-056 (comunidades, régimen, embebido) | Sí | T002, T006, T007, T012 | Ver C1 |
| FR-060–FR-069, FR-102 (skill) | Sí | T016, T019–T021 | |
| FR-070–FR-074 (leyes vertebrales) | Sí | T013, T015, T016 | |
| FR-080–FR-085 (evals) | Sí | T017–T019, T022, T025 | |
| FR-086 (pausas [datos]) | Sí | T001–T003, T005, T010, T012, T014, T015 | |
| FR-090–FR-099 (tests, calidad, docs) | Sí | T008, T010–T012, T023, T024 + batería | |
| FR-100, FR-101 (aceptación) | Sí | T021, T025 | |
| SC-001–SC-015 | Sí | Tabla de trazabilidad de tasks.md | Todos con al menos una tarea |

Los tests nombrados en `plan.md` aparecen en `tasks.md`, salvo seis que son controles ya existentes que el hito solo mantiene en verde (`TestDependenciasDelBinario`, `TestElBinarioNoEnlazaLosEjemplos`, `TestEsquemaDeEval`, `TestGramaticasCoincidenConBoe`, `TestLasFuentesNoFirmanComoKitlegal`, `TestRegistroDeProduccion`); no es un hueco.

## Alineación con la constitución

Sin conflictos. Comprobado: skills primero y evals antes que la skill (T019 < T021), genericidad territorial (ningún municipio en código ni datos; T006, T007, T019), frontera de red (ninguna tarea abre red; solo T025 usa `gh` y `git push`), fusionar es humano, sin atajos ni ADR nuevo (justificado).

## Tareas sin requisito

Ninguna. T022–T025 se ligan a FR-083, FR-098/099, FR-093–FR-097 y FR-100/101.

## Métricas

- Requisitos: 79 FR + 15 SC = 94
- Tareas: 25 (10 `[datos]`, 1 `[plataforma]`)
- Cobertura (requisitos con ≥ 1 tarea): 100 %
- Ambigüedades: 1 (C1)
- Duplicidades: 0
- Marcadores sin resolver (TODO, NEEDS CLARIFICATION): 0
- Incidencias CRÍTICAS: 0 · ALTAS: 0 · MEDIAS: 2 · BAJAS: 3

## Siguientes acciones

Sin incidencias críticas ni altas: se puede pasar a `/speckit-implement`. Antes, conviene resolver C1 en `tasks.md` (una frase en T002) y valorar C2. C3 y C4 se verifican al llegar a T019 y T012.

## Resolución de los hallazgos (corrección posterior al juicio de `tasks`)

El juez de `tasks` rechazó la consistencia por un bloqueante de quickstart y añadió tres observaciones. Lo aplicado, sobre los artefactos y no sobre este informe:

| ID | Estado | Dónde |
|----|--------|-------|
| Bloqueante del juez: el caso «código bien formado que no está en la relación» ejercido con un código de provincia fuera de `01`-`52` | Resuelto como **clase de error**, en todos los artefactos | `quickstart.md` §6 (el código ausente se deriva del fichero congelado y se comprueba antes de invocar; `99999` queda como caso explícito de exit 2), `tasks.md` T009 y T012, `plan.md` (`TestResolver/codigo-inexistente` y control 6), `spec.md` (US4 escenario 4, FR-011 y FR-012), `data-model.md` §2.6, `contracts/applet-territorio.md` §exit codes y `research.md` D9 |
| C1 (MEDIA) | Resuelto | `tasks.md` T002: origen de cada dato (código y provincias desde `municipios.yaml`, ya congelado; nombre, régimen y BOCM aportados y revisados por la persona en la pausa) y parada con anotación en `gates/tarea-T002.md` |
| C2 (MEDIA) | Resuelto partiendo la tarea | El antiguo T016 se parte en **T016** (lector de la jerarquía, su control y `Makefile`, verde por sí sola) y **T017** (normas nuevas, marcas `vertebral` y tabla de generadores, indivisibles por `TestRegenerarYComparar` y `TestNormasDelRepositorio/vertebrales`). Las tareas siguientes se renumeran: el hito pasa de 25 a **26** tareas y el paso 16 del plan se declara divisible en `plan.md` |
| C3 (BAJA) | Infundada, según comprobó el juez | Ningún control exige una carpeta de skill por carpeta de evals; T020 (antes T019) puede crear `evals/legal-core/` antes que la skill |
| C4 (BAJA) | Resuelta | El umbral de 200 ms de T012 es el ya establecido en `internal/app/testdata/script/boe-cache-rapida.txtar` |
| C5 (BAJA) | Resuelta | `tasks.md`, batería: las descargas se atribuyen a las pausas de T001, T003 y T015; la de T002 no descarga nada |
| Observación del juez sobre SC-003 | Resuelta | `tasks.md`, tabla de trazabilidad: la fila SC-003 incluye T007, que es donde vive el subtest `regimen-de-todas` |

Con la división de C2, las métricas de arriba quedan en **26 tareas** (10 `[datos]`, 1 `[plataforma]`); el resto de los recuentos no cambia.
