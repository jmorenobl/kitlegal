# Specification Analysis Report — H23 · skill `jurisprudencia` + `cita preparar` y `cita cotejar`

Análisis de solo lectura de `spec.md` (55 FR, 10 SC, 6 historias, 29 escenarios), `plan.md`, `tasks.md` (T001 a T012) y, como apoyo, `research.md`, `contracts/` y `gates/supuestos.md`, contra la constitución 2.13.0. Rama `019-h23-skill-jurisprudencia-ninguna`.

## Lo que se ha comprobado en esta sesión, y lo que no

Comprobado (leído o ejecutado y visto terminar):

- `evidencias/adr-0036/ecli-es-ts-2023-3144-fragmento.txt`: 39 líneas, 2 353 bytes, SHA-256 `4886e0c8…ee27`, como dicen spec (SC-008), plan y la cabecera de `tasks.md`.
- Las 28 rutas de código, esquemas, flujo y documentos que nombran T002 a T011 existen, y los 19 guiones e2e de T009 (17 de `internal/app/testdata/script/` y 2 de `internal/skills/testdata/script/`) existen con esos nombres.
- El arnés de e2e tiene `arbol`, `mcp` y `KITLEGAL_T0_BIN`, que T001 usa; `TestOrdenesDeLasSkillsEmpotradas` existe en `internal/app/skills_test.go`.
- `docs/JURISPRUDENCIA.md` trae el ECLI `ECLI:ES:TC:2024:79`, la ficha del ROJ `STS 3144/2023`, la dirección del buscador y las casillas que el spec cita; no trae ninguna dirección del buscador del Tribunal Constitucional.
- `checklists/requirements.md`: 0 ítems sin marcar. Ningún marcador pendiente en spec, plan ni tasks.
- Aritmética de plan y contratos: 6 evals × 3 repeticiones = 18 respuestas por modo; 485 + 73 × 272 = 20 341 s, bajo los 21 120 s (352 min).

No comprobado: no se ha ejecutado ningún test ni `make ci`; no se ha re-medido el prototipo de `research.md` (V21, M4: los 11 tests y los guiones que rompe la llegada del applet y de la skill) ni se ha mirado el contenido de `/tmp/claude-501/kl-proto`; no se ha comprobado que cada guion nombrado cambie solo en una lista literal. Esas cifras se toman de `research.md` tal como están.

## Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| A1 | Ambigüedad / control | MEDIUM | spec.md FR-047, Assumptions («La dirección del buscador del Tribunal Constitucional»); contracts/evals-jurisprudencia.md §5 eval 05; tasks.md T007, T009 | La dirección `https://hj.tribunalconstitucional.es/` no está en ningún fichero del repositorio; la fija el plan de memoria (research S1) y entra igual en el paso 7 de `SKILL.md` (T009) y en lo que espera la eval 05 (T007). La eval compara la respuesta con la dirección que la propia skill da: si la dirección es falsa, la eval pasa igual. | Dejar la comprobación de la dirección como pendiente de la persona en el informe final, junto a su línea de `gates/supuestos.md` (ya está); no tratar la eval 05 como prueba de que la dirección es la correcta. La constitución (II, «nunca se inventa contenido legal») no se viola mientras el informe lo diga. |
| C1 | Cobertura de control | MEDIUM | spec.md FR-042 (tercera viñeta), FR-043, FR-046, FR-049, US3.7; `gates/supuestos.md`; tasks.md T012 (5) | FR-065 y `supuestos.md` anotan como no comprobados solo «resume» y «no se corresponden» (y la línea de FR-047). Otras reglas de la skill que el spec pide con MUST no tienen eval ni control y no están en esa lista: no preparar como ROJ un número sin fecha (US3.7: ninguna de las seis evals lo ejerce), que la respuesta diga que el documento no es el pedido (FR-051 solo compara la línea, la ausencia de cita con ese ROJ y las operaciones), no decir que una sentencia existe o no existe (en «Fuera de alcance» del spec, no en los supuestos), y las reglas de FR-049. El informe final solo enseña como no comprobado lo que está en `supuestos.md`. | Añadir a `gates/supuestos.md`, como `[alcance]` o `[skill]`, una línea por cada una de esas reglas «de la skill sin control en este hito», para que el informe final no las dé por comprobadas. T012 (5) puede constatar que siguen ahí. No hace falta ninguna eval nueva: el spec las deja fuera a propósito. |
| I1 | Inconsistencia | LOW | plan.md («Scale/Scope», «Tests y guiones existentes que cambian», *Complexity Tracking* fila 3); research M4; tasks.md (cabecera «Diecinueve guiones») y T009; `supuestos.md` | El plan dice 11 tests y 21 guiones (4 por el applet, 17 por la skill); `tasks.md` y T009 alinean 23 (4 y 19), porque `test-integration` rompe además `instalar.txtar` e `instalar-sin-gobin.txtar`. Está declarado en `supuestos.md` y `plan.md` no se ha cambiado. Las rutas de T009, que son las que vigila el guardián, sí son las 19. | Sin acción para el run: manda `tasks.md`. Quien lea `plan.md` debe saber que dos cifras van por debajo. |
| I2 | Inconsistencia de términos | LOW | spec.md FR-044; contracts/SKILL-jurisprudencia.prototipo.md (paso de la cita, «las siglas») | FR-044 dice que la cita lleva «el órgano, el número de resolución y la fecha» y que todos sus datos son los de la ficha; el ejemplo es `STS 1088/2023, de 4 de julio`, y la ficha trae el órgano como «Tribunal Supremo. Sala de lo Civil». El borrador de la skill escribe «las siglas». De dónde salen «STS» (las siglas del ROJ de la ficha) no está en el spec. Ningún control compara lo que va antes del corchete (FR-051). | Anotar que la parte anterior al corchete no se compara y que el borrador toma las siglas del ROJ; para otros órganos la forma no está definida y no se cubre en el hito. |
| I3 | Inconsistencia de términos | LOW | spec.md FR-030, US4.3; plan.md D14 y *Complexity Tracking* fila 2; tasks.md T005 | El spec pide que las herramientas lleven «los esquemas de `--describe` de su verbo»; el plan añade `x-banderas` a `--describe` y la quita del esquema de la herramienta, así que el de la herramienta es el de `--describe` menos esa anotación. T005 lo comprueba con el auxiliar que quita las banderas globales y con `TestEsquemasDeHerramienta`. | Sin acción: el cambio del plan está justificado y comprobado. |
| U1 | Alcance fuera del spec | LOW | tasks.md T004, T005; plan.md *Complexity Tracking* filas 1 y 2 | `Registro.LeerDe`, `Despacho.Entrada`, `x-banderas` y el cambio de `schemas/instalacion.json` no están en el spec; los exigen FR-020 y FR-030 (entrada que no lea una llamada de herramienta, sin tocar `mcp`) y FR-040 (tabla generada con cada orden). Cada uno se traza a su requisito y la alternativa más simple se rechaza con motivo. | Sin acción. Encaja con la regla de proporcionalidad (cada mecanismo se traza a un requisito). |

## Tabla de cobertura

Los 55 requisitos y los 10 criterios de éxito están en «Requisitos → tareas» de `tasks.md`.

| Requisito | ¿Tarea? | Tareas | Notas |
|---|---|---|---|
| FR-001 a FR-006 (applet, sin red, sin caché ni grafo, sobre, referencia) | Sí | T001, T002, T003, T006 | Guion `cita-sin-efectos` y `TestArquitectura` |
| FR-010 a FR-015 (`cita preparar`) | Sí | T001, T003, T006 | Golden y e2e |
| FR-020 a FR-026 (`cita cotejar`) | Sí | T001, T003, T004, T006 | La entrada la entrega el kernel (T004) |
| FR-030, FR-031 (herramientas, plugin) | Sí | T004, T005, T006; FR-031 es una restricción de todas y la constata T012 | `TestHerramientasDelServidor` |
| FR-040 a FR-049 (skill) | Sí | T005, T009 | FR-041 y FR-044 los decide el umbral de T008; FR-045 y FR-048, sin control (FR-065); ver C1 |
| FR-050 a FR-056 (evals) | Sí | T007, T010 | `TestPreguntasConElFragmento` para FR-053 |
| FR-060 a FR-063 (umbrales y job) | Sí | T008, T010 | Cuatro umbrales con `decide: true` |
| FR-064 | Restricción | Batería de verificación; T008; T012 | No es una tarea; la constata T012 |
| FR-065 | Supuestos | Ya en `gates/supuestos.md`; T012 | Ver C1 |
| FR-066 | Plan | «Controles de umbral» de `plan.md` | Una fila por umbral que decide y por cada control de `make ci` |
| FR-070 | Sí | T011 | |
| FR-080 a FR-088 (controles y DoD) | Sí | T001, T002, T003, T006, T007, T008, T009, T012 | |
| SC-001 | Cierre del workflow | No es una tarea | Con los controles de T008, T010 y la skill de T009 |
| SC-002 | Persona, después del run | No es una tarea | |
| SC-003 a SC-009 | Sí | T001, T006, T007, T008, T009, T012 | |
| SC-010 | Revisión final | T011 | |

## Alineación con la constitución

Sin conflictos. Comprobado contra los principios:

- **I (frontera humana, CENDOJ)**: el applet no alcanza `net`, `net/http` ni `internal/httpx` (T006, `TestArquitectura`); ninguna tarea toca la red ni graba nada de una fuente (plan, «Datos externos»).
- **II (cita y fuente)**: el sobre dice que no hay consulta (`fuente` `kitlegal.cita`); la reserva es A1.
- **III (tests primero)**: T001 escribe cuatro guiones congelados antes de cualquier código; las evals entran con T007, antes de la skill (T009).
- **V (YAGNI)**: ninguna dependencia nueva; las desviaciones están en *Complexity Tracking*.
- **VIII (skills primero)**: la skill y sus seis evals, y las dos herramientas que ella usa.
- **Umbrales (ADR 0029, 0037)**: los cuatro umbrales que deciden tienen control con forma en «Controles de umbral» y un test que los ve fallar (T008, T010). `jurisprudencia` no tiene juez (FR-055).
- **Modo desatendido**: T001 es la única `[aceptacion]` y va primera; las seis tareas que tocan `schemas/` o `testdata/` llevan `[datos]`; ninguna tarea nombra la carpeta de evidencias por su ruta ni pide nada a una persona.

## Tareas sin requisito

Ninguna. T004 y T005 son anteriores al requisito que las exige (FR-020, FR-030, FR-040); ver U1.

## Métricas

- Requisitos: 55 FR y 10 SC (los SC con trabajo construible: SC-003 a SC-009)
- Tareas: 12 (T001 a T012)
- Cobertura de requisitos con al menos una tarea o control explícito: 100 % (los que no tienen tarea —FR-031, FR-064, FR-065, FR-066, SC-001, SC-002 y SC-010— tienen su control o su constancia nombrados)
- Ambigüedades: 1 (A1)
- Duplicados: 0
- Inconsistencias: 3 (I1, I2, I3)
- Huecos de control sin anotar: 1 (C1)
- Problemas críticos: 0

## Próximas acciones

- No hay problemas críticos: se puede seguir con la implementación.
- Los dos hallazgos de severidad media se resuelven con texto en `gates/supuestos.md` y en el informe final, sin cambiar spec, plan ni tareas: A1 (la dirección del Tribunal Constitucional, que la persona comprueba) y C1 (las reglas de la skill sin control, que el informe no debe dar por comprobadas).
- Los tres de severidad baja (I1, I2, I3) y U1 no cambian lo que el run hace.
