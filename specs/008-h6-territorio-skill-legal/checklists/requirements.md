# Specification Quality Checklist: H6 · `territorio` + skill `legal-core` v0

**Purpose**: Validar la completitud y calidad del spec antes de pasar a planificación
**Created**: 2026-09-20
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Success criteria are technology-agnostic (no implementation details)
- [x] All acceptance scenarios are defined
- [x] Edge cases are identified
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria
- [x] User scenarios cover primary flows
- [x] Feature meets measurable outcomes defined in Success Criteria
- [x] No implementation details leak into specification

## Notas de validación

- **«No implementation details» y «No implementation details leak»**: el spec nombra solo lo que nombra el propio hito (`internal/core/ids`, `data/territorio/`, `data/normas.yaml`, `references/leyes_vertebrales.md`, `jerarquia_normativa.md`, `cobertura`, la tarea `[datos]`, `boe buscar`) y las piezas que la Definition of Done y la constitución imponen a cualquier hito (`schemas/`, `--describe`, `make ci`, e2e con `testscript`, `testdata/fuzz/`, `CHANGELOG.md`). FR-004 fija `fuente` y `url` del sobre porque no es diseño libre: ADR 0006 ya decidió que un applet que no consulta fuentes emite el espacio reservado `kitlegal.` / `kitlegal:`, y de ahí depende una regla visible para quien lee la salida —que ese resultado no es una cita de fuente pública—. No se fija ningún formato de fichero de `data/territorio/`, ni el algoritmo de coincidencia por nombre, ni la estructura interna de `cobertura`, ni cómo se genera cada referencia: eso es del plan. Tampoco se fija **con qué mecanismo viajan los datos dentro del binario ni en qué paquete vive ese empaquetado**: FR-056 quedó reescrito en términos observables (la respuesta no depende del directorio de trabajo ni de ninguna ruta configurable; un fichero ausente o inválido se detecta antes de ejecutar y no en ejecución, de donde se sostiene FR-016), y el mecanismo (`go:embed`) y la colocación del paquete viven en `plan.md`; el rastro de la decisión queda en *Clarifications* (Q4), marcado como detalle para `plan`. Lo mismo con el juicio mecánico de las evals: FR-084 exige que el formato común se extienda de forma compatible hacia atrás y deja a `plan.md` qué paquete lo implementa.
- **«Focused on user value»**: US1 y US2 son la misma pregunta de la aceptación dentro y fuera del territorio configurado; US3 y US4, lo que impide razonar mal (régimen foral, ambigüedad); US5, la skill que es el producto; US6 y US7, lo que sostiene que el dato sea dato y no memoria.
- **«All mandatory sections completed»**: *Resumen*, *Criterios del hito, literales* (con la trazabilidad), *User Scenarios & Testing* (US1 a US7 y *Edge Cases*), *Requirements* (FR-001 a FR-102, incluidos FR-085 y FR-086, y *Key Entities*), *Success Criteria* (SC-001 a SC-015), *Fuera de alcance* y *Assumptions*.
- **«No [NEEDS CLARIFICATION] markers remain»: se cumple.** El spec se escribió en modo desatendido con exactamente 3 marcadores —los que no se podían resolver con el hito, `CLAUDE.md`, `refs/` ni la constitución delante—, y los tres se resolvieron en la sesión de `clarify` del 2026-09-20, junto con dos decisiones más que la misma sesión cerró. Hoy el spec tiene **0 marcadores**. Dónde aterrizó cada respuesta:
  1. **FR-048** (qué municipios entran en el fichero INE→DIR3) → Q1/B: entra todo municipio con fila en el REL cuyo número de inscripción es coherente con su código INE y su dígito de control; la muestra verifica la regla, no las filas, y su evidencia va al registro de FR-047; lo que queda fuera se declara en `cobertura` (**FR-048** y **FR-023**), y el tercer estado por fila queda en *Fuera de alcance*.
  2. **FR-067** (de dónde sale en `data/` el contenido de las dos referencias generadas) → Q2/B: campo booleano opcional `vertebral: true` en `data/normas.yaml`, con su ampliación de `schemas/normas.yaml.json`, y un fichero de datos nuevo con esquema propio para la jerarquía (**FR-067** y **FR-070**, más *Key Entities*).
  3. **FR-084** (cómo se expresan y se juzgan las evals de `legal-core`) → Q3/A: el formato común se extiende en este hito de forma compatible hacia atrás —variante de `comandos` para territorio, esperado propio de territorio y «al menos un esperado verificable» en lugar de `citas` obligatorias—, sin tocar las evals de `boe-legislacion` (**FR-084**).
  4. **FR-056** (cómo obtiene el binario los ficheros de `data/territorio/`) → Q4/A: viajan con el binario; la respuesta no depende del directorio de trabajo ni de ninguna ruta, y la falta de un fichero se detecta antes de ejecutar, de donde se sostiene **FR-016**.
  5. **FR-069** (qué hace `legal-core` cuando la pregunta exige el texto de una norma) → Q5/A: delegación en un solo sentido hacia `boe-legislacion`, tabla de comandos solo con los verbos de `territorio` y evals de territorio, cobertura y no activación (**FR-069**).
  Las cinco entradas están en *Clarifications · Session 2026-09-20* con el prefijo «(auto: criterio …; fuente: …)» que exige la constitución, y las alternativas rechazadas quedan en `gates/clarify-respuestas.json`.
- **«Requirements are testable and unambiguous»**: cada FR tiene escenario o criterio medible. FR-001 a FR-009 → US1 escenarios 1 a 4 y SC-001; FR-010 a FR-016 → US4 escenarios 1 a 4 y SC-004; FR-020 a FR-022 → US2 escenarios 2 a 4 y SC-002; FR-023 y FR-048 → US7 escenario 3 y SC-008; FR-024 → SC-005 (ningún fichero fuera de `data/territorio/`) y Definition of Done §11; FR-030 a FR-035 → US4 escenario 3 y SC-006; FR-040 a FR-045 y FR-049 → US7 escenarios 1 y 4 y SC-007, SC-014; FR-046 y FR-047 → US7 escenario 2 y SC-008; FR-050 a FR-055 → US1 escenario 2, US2 escenario 1, US3 escenarios 1 a 3, SC-003 y SC-005; FR-056 → US1 escenarios 3 y 4 y SC-001 (mismo `data` y misma huella por nombre y por código, salida byte a byte idéntica con `--offline`, sin dependencia del directorio de trabajo) más SC-012 (`make ci` en verde valida los ficheros contra su esquema antes de ejecutar); FR-060 a FR-068 → US5 escenarios 1 a 4 y SC-009; FR-069 → US5 escenarios 1 y 2 y SC-009 (tabla de comandos sin drift, solo con los verbos de `territorio`) y SC-011 (las evals miden territorio, cobertura y no activación); FR-070 a FR-074 → US6 escenarios 1 a 3 y SC-010, más SC-015 para el verde del conjunto de `boe-legislacion`; FR-080 a FR-084 → US5, SC-011 y SC-015; FR-085 y FR-086 → SC-007 y SC-014 en su parte de datos y esquemas, y la comprobación es la del propio workflow (toda tarea que toca `schemas/` o `testdata/` lleva `[datos]`, con pausa humana ante material existente); FR-090 a FR-097 → SC-012; FR-098 y FR-099 → SC-014; FR-100 a FR-102 → US5 escenarios 1 y 2 y SC-013.
- **«Success criteria are measurable» y «technology-agnostic»**: SC-001 a SC-015 se cuentan o se comprueban sobre el repositorio, sobre la salida y sobre el informe del job de evals (8 elementos de `data`, igualdad byte a byte con `--offline`, N candidatos de un nombre compartido, régimen en el 100 % de los municipios, 3 casos de la muestra de verificación, un único fichero de comunidad relleno, diff vacío de lo generado, umbrales de cobertura, ninguna conexión abierta, y en SC-015 la serie de 3 sesiones con umbral de 2 de ADR 0016, con el commit y el id del modelo) sin fijar lenguaje, biblioteca ni diseño interno. Los identificadores de nivel de calidad que aparecen (85 % y 70 %) son los de la Definition of Done §1.9.
- **«All acceptance scenarios are defined»**: los cuatro casos de la matriz territorial del control están en US1 (cubierto), US2 (no cubierto, nada inventado), US3 (foral) y US4 (ambiguo); el fuzz, en FR-035 y SC-006; la muestra de la tarea `[datos]`, en US7 escenario 2 y SC-008; las evals, en US5, SC-011 (composición y orden) y SC-015 (verde del conjunto, con la regla de ADR 0016); la aceptación, en FR-100 a FR-102 y SC-013.
- **«Edge cases are identified»**: nombre oficial del INE con artículo pospuesto y nombre bilingüe, mayúsculas y diacríticos, entrada de solo cifras que no es código, código con y sin dígito de control, ciudades autónomas, comunidad uniprovincial distinta de Madrid, municipio sin DIR3 verificado, municipio desaparecido por fusión, cobertura con un aspecto incompleto y régimen foral en los territorios históricos frente a Navarra.
- **«Scope is clearly bounded»**: *Fuera de alcance* enumera lo que el hito no pide y lo que pertenece a otros hitos ya numerados (festivos y plazos de H9, grafo de H7, `.kitlegal/config.yaml` y asunto de H10, `data/boletines/` de H14, verticales del backlog, otros territorios del backlog), más lo que la constitución prohíbe (pedir DIR3, REL o INE por red; sortear un WAF o la Red SARA).
- **«Dependencies and assumptions identified»**: *Assumptions* recoge la procedencia del sobre por ADR 0006, el régimen como dato nacional, la relación del INE como registro, la derivación del DIR3 como hipótesis, la verificación manual de la muestra, la tabla de leyes vertebrales de `refs/`, los municipios de prueba, las dependencias de H1 a H5 y la forma de la ejecución de aceptación.
- **«All functional requirements have clear acceptance criteria»**: también los que no son comportamiento del producto: FR-024 (SC-005 y Definition of Done §11), FR-045 y FR-073 (tarea `[datos]` con pausa, SC-007 y SC-010), FR-047 (registro de la muestra, SC-008), FR-049 (SC-014, `docs/SOURCES.md` cierto y `verify-sources.sh` sin casos nuevos), FR-083 (SC-011 para el orden de commits y SC-015 para el verde), FR-085 y FR-086 (régimen de capa 3: la etiqueta `[datos]` y la pausa las comprueba el propio workflow sobre el diff de cada tarea) y FR-098 y FR-099 (SC-014).
- **«User scenarios cover primary flows»** y **«Feature meets measurable outcomes»**: ver la tabla de trazabilidad de *Criterios del hito, literales*, que lleva cada criterio literal del hito a sus FR, US y SC.

## Notas

- Los ítems marcados incompletos exigen actualizar el spec antes de `/speckit-plan`. Tras la sesión de `clarify` del 2026-09-20 y la corrección del gate del spec, no queda ninguno: los tres marcadores están resueltos en el spec y en *Clarifications*, y FR-056 y FR-084 ya no fijan mecanismo ni paquete.
