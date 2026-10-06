# Specification Analysis Report — H23 (`018-h23-cita-resolver-comprobar`)

Análisis de `spec.md`, `plan.md` y `tasks.md` (con `research.md`, `data-model.md` y `contracts/` como apoyo) contra la
constitución 2.12.0. Solo lectura del árbol: se leyeron ficheros y se hicieron `grep` y `find`; **no se ejecutó ningún
test, `make ci`, `go build` ni consulta de red**. Lo que dice el informe sobre el comportamiento de `make ci` es
lectura del código de los tests, no una ejecución.

Hooks de extensión: `before_analyze` y `after_analyze` solo declaran `speckit.git.commit`, opcional. No se ejecutó: la
sesión es de un paso y el commit lo hace el workflow.

## Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| C1 | Cobertura / Inconsistencia | HIGH | tasks.md T021; `internal/evals/conjunto_test.go:1559-1620` (`probarLineaSinConsulta`) | La subprueba `linea-sin-consulta` de `TestEvalsDelRepositorio` recorre **toda** skill de `skills/` (`skills.Listar`) y exige en su `SKILL.md`, en un bloque `text`, la línea exacta `⚠ SIN CONSULTA AL BOE: <causa>. Para consultarlo hace falta instalar kitlegal: https://kitlegal.es/instalar/`. El prototipo de `jurisprudencia` (`contracts/skill-jurisprudencia.prototipo.md:156`) dice qué hacer sin binario con `⚠ SENTENCIA NO COMPROBADA:` y no lleva esa línea; ni spec, ni plan, ni research, ni tasks la nombran (`grep -rn 'SIN CONSULTA'` en el directorio del hito: 0 coincidencias). Al añadir la skill, T021 deja `make ci` en rojo. T021 declara `internal/evals/conjunto_test.go` en sus rutas, así que puede arreglarse en la tarea, pero el spec no decide cómo y las dos salidas tocan algo fijado: meter en una skill del CENDOJ una línea que habla del BOE contradice FR-063 y FR-068; eximir a `jurisprudencia` del test cambia una regla de H21. | Decidir antes de implementar y dejarlo escrito en spec o plan: o `skillsConLineaSinConsulta` pasa a ser «las skills de la tabla que consultan el BOE» y el test se acota por una propiedad de la skill (no por nombre), o la línea del test se generaliza a una forma con la fuente como parámetro. Declararlo en T021 (tests que cambian) y en «Tests existentes que cambian» del plan. |
| C2 | Cobertura | HIGH | tasks.md T021 y «Complexity Tracking» fila 3; research V35; `internal/skills/testdata/script/instalar.txtar:28,33`, `instalar-sin-gobin.txtar:19`; `internal/skills/instalacion_test.go:73` | V35 contó los guiones afectados con `grep -c "boe-legislacion.*legal-core"` (las dos skills en la misma línea) y dio 15 guiones de `internal/app/testdata/script/`. `TestInstalacion` (`//go:build integration`, parte de `make ci` por `test-integration`) copia `skills/` **entera** a `$WORK/repo`, ejecuta `make install` y compara con `\A…\z` el listado exacto: `exec ls -A $HOME/.agents/skills` → `boe-legislacion\nkitlegal.json\nlegal-core\n` y `exec ls -A $HOME/.claude/skills` → `boe-legislacion\nlegal-core\n`. Con `skills/jurisprudencia/` esas tres aserciones fallan. Ninguna de las dos rutas (`internal/skills/testdata/script/instalar.txtar`, `instalar-sin-gobin.txtar`) está en las rutas de T021, ni V35 ni el plan las mencionan. `instalar-de-nuevo.txtar` solo usa aserciones por línea y no se rompe. | Añadir los dos guiones a las rutas de T021 (es `[datos]`) y a la cuenta de «guiones de hitos anteriores que cambian» (17, no 15, más `argumentos` y `h21-mcp-herramientas`). Hoy la cláusula de T021 «cualquier otra enumeración… redelimitando la tarea» lo recogería, pero a costa de un intento fallido y una redelimitación. |
| A1 | Ambigüedad | MEDIUM | spec FR-010, FR-022; plan «Datos externos»; research S1; tasks T011 | La estructura HTML del buscador no se ha visto (ninguna grabación en el repositorio; el plan y el spec lo dicen). FR-010 y FR-022 hacen depender toda la lectura de que de cada resultado salgan ECLI, ROJ, fecha y URL con expresiones regulares; si no salen, toda la respuesta es «no reconocida» (5). T011 prevé qué hacer si las grabaciones contradicen un supuesto (cambia la lectura, no la grabación), pero no hay forma de saber antes del run si los ocho campos son legibles sin biblioteca de HTML. | Sin cambio de artefactos: es el riesgo declarado del hito. El informe final tiene que enseñar S1 como supuesto cerrado o no. |
| A2 | Ambigüedad | LOW | spec «Assumptions» (cotas del ECLI, dirección del buscador del TC, operadores del buscador); `gates/supuestos.md` | Tres datos escritos sin comprobar (órgano ≤ 7 y número ≤ 25 del ECLI; `https://hj.tribunalconstitucional.es/`; operadores del buscador). Están anotados como supuestos y la eval (e) toma la dirección de `SKILL.md`, así que un error no se detecta por ningún control. | Los lee la persona en el informe final; ya está dicho. |
| I1 | Inconsistencia | LOW | spec US3, escenarios 4, 6, 5 | La lista de aceptación de US3 está numerada 1, 2, 3, 4, **6**, **5**: el escenario de «más de una resolución» lleva el 6 y va antes del 5. | Renumerar. Sin efecto en tareas: nadie cita el número. |
| U1 | Subespecificación | LOW | spec (sin FR); plan D16; tasks T008 | El cambio de `--describe` (`title` en las banderas propias de todo verbo) y de `schemas/instalacion.json`, un fichero publicado de H19, no tiene requisito propio en el spec: T008 lo traza a FR-001, FR-060 y FR-069. Está justificado en *Complexity Tracking* y en `gates/supuestos.md` ([comportamiento]) y es revisión de capa 3. | Sin cambio; mantenerlo visible en el informe final. |

Sin hallazgos de duplicación. Sin hallazgos de terminología dispersa que afecte a la ejecución.

## Coverage Summary

Todos los requisitos tienen tarea o una razón escrita para no tenerla. Mapa completo en «Trazabilidad» de `tasks.md`; lo
que no es una tarea, por diseño:

| Clave | ¿Tarea? | Tareas | Notas |
|---|---|---|---|
| FR-001 a FR-006 | Sí | T001, T008, T009, T014, T015 | |
| FR-010 a FR-019 | Sí | T001, T011 a T015 | FR-013 y FR-014 solo con listas sintéticas y las grabaciones de una resolución (T012) |
| FR-020 a FR-024 | Sí | T005, T007, T009, T012, T014 | FR-024 sin control nuevo, por el propio requisito |
| FR-030 a FR-034 | Sí | T005, T006 | |
| FR-040 a FR-044 | Sí | T013, T014, T019 | |
| FR-050, FR-051 | Sí | T015 | |
| FR-060 a FR-068 | Sí | T021 (+T008) | FR-061, FR-064 y FR-067 sin control propio, por el spec |
| FR-069, FR-076 | Restricción | batería, T021, T023 | los comprueba el diff |
| FR-070 a FR-075 | Sí | T017 a T021 | |
| FR-080 a FR-083 | Sí | T018, T020, T021 | |
| FR-084 | Sí | plan «Controles de umbral», T023 | 4 filas que deciden + 1 de `boe-legislacion` + 17 de `make ci` |
| FR-085 | Sí | `gates/supuestos.md` ya escrito; T023 lo comprueba | |
| FR-090 a FR-095 | Sí | T010, T011, T016, T022 | |
| FR-100 | Sí | T022 | |
| FR-110 a FR-118 | Sí | T001 a T023 | |
| SC-002 a SC-007, SC-009 | Sí | T001, T005 a T008, T012, T014 a T016, T020, T021, T023 | |
| SC-001 | No (es el cierre del workflow) | T020, T021 construyen los controles | por diseño |
| SC-008 | No (humano, tras el run) | — | por diseño |

## Constitution Alignment Issues

Ninguna violación.

- **Principio I**: la excepción del `POST` coincide con el texto de la 2.12.0 (formulario declarado por la fila revisada de `docs/SOURCES.md`, emitido solo por `internal/httpx`, a la dirección declarada). La fila existe en `main` (`docs/SOURCES.md:43`, revisada 2026-10-03, ritmo `5s`, formulario `POST /search/search.action`, campos `ECLI`, `ROJ`, `NUMERORESOLUCION`).
- **Principio V**: ninguna dependencia nueva; `httpxtest` es la única estructura nueva fuera de la arquitectura objetivo y está en *Complexity Tracking*.
- **Capa 1 / «Controles de umbral»**: el plan tiene la sección, con una fila por cada umbral que decide (`cita_sin_resolver` y `sin_activar`, en los dos modos) y un control con forma en cada una; T020 y T021 los construyen.
- **Aceptación primero**: T001 es la única tarea `[aceptacion]` y es la primera.
- **ADR 0037 / FR-076**: no se crea `evals/jurisprudencia/juez/`; «resume» queda anotado como no comprobado.

## Unmapped Tasks

Ninguna: las 23 tareas citan FR o SC. T023 cita los requisitos de la Definition of Done que cierra.

## Orden y dependencias

La tabla de dependencias no tiene ninguna arista hacia delante. Comprobado a mano: T007 (sitio de prueba) antes de T012 que lo usa; T008 antes de T015 (`title` en `--describe` antes del primer verbo con banderas); T010 antes de la grabación y T011 después; T015 antes de T018 (herramienta `cita_resolver` en el registro). Las rutas bajo `testdata/` y `schemas/` están en tareas `[datos]` (T003, T008, T010, T011, T015, T017, T021).

## Metrics

- Requisitos: 72 FR + 9 SC = 81
- Tareas: 23
- Cobertura (requisitos con ≥ 1 tarea): FR 72/72 = 100 %; con SC, 79/81 = 97,5 % (SC-001 y SC-008 son del cierre y de una persona, por diseño)
- Ambigüedades: 2 (A1, A2)
- Duplicaciones: 0
- Hallazgos CRITICAL: 0 · HIGH: 2 (C1, C2) · MEDIUM: 1 (A1) · LOW: 3 (A2, I1, U1)

No medido en esta sesión: cualquier cifra de `make ci` (tiempos, cobertura, que los guiones congelados fallen hoy por su precondición), el peor caso de duración del job de evals (20 613 s frente a los 352 min del tope, cálculo del plan, S9) y el comportamiento real de las respuestas del CENDOJ.

## Next Actions

- Hay dos HIGH, ambos con la misma forma: una enumeración de skills o de ficheros en un test existente que la skill nueva rompe y que el plan no recogió. No son bloqueantes del diseño, pero T021 los encontraría a mitad de la tarea. Resolverlos antes de `/speckit-implement`:
  - C1: decidir en el plan cómo `linea-sin-consulta` trata a `jurisprudencia` y escribirlo en «Tests existentes que cambian» y en T021.
  - C2: añadir `internal/skills/testdata/script/instalar.txtar` e `internal/skills/testdata/script/instalar-sin-gobin.txtar` a las rutas de T021 y corregir la cuenta de guiones de V35 y *Complexity Tracking* (de 15 a 17 en total, repartidos en dos directorios).
- A1 y A2 los arrastra el informe final como supuestos; I1 y U1 pueden esperar.
