# Specification Analysis Report — H7.1 (`011-h7-1-graph-check-acotado`)

Solo lectura: `spec.md`, `plan.md`, `tasks.md` (con `research.md`, `data-model.md`, `contracts/`, `quickstart.md` y `gates/supuestos.md` como apoyo) contra la constitución 2.6.0. Contrastado con el repositorio: guiones `internal/app/testdata/script/h7-*.txtar`, `internal/graph/lectura.go` y las tablas de `skills/*/SKILL.md`. No se ha modificado ningún otro fichero.

**Resultado: sin hallazgos CRÍTICOS ni ALTOS.** Se puede pasar a `/speckit-implement`. Hay una inconsistencia MEDIA que conviene corregir antes (A1) y cuatro observaciones BAJAS.

## Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| A1 | Inconsistencia | MEDIA | tasks.md T025 (l. 180) frente a spec.md FR-080 (l. 275-281), plan.md l. 165 y l. 271, y T002/T014/T017 | T025 habla de «los seis guiones de H7 de FR-080» y de que «ningún guion del arnés fuera de los seis de FR-080 ha cambiado». FR-080 tiene seis viñetas pero **ocho ficheros** (`codigos`, `entrega-fallida`, `concurrencia`, `version-obsoleta`, `fuente-caducada`, `no-emiten`, `applet`, `show`), y el plan dice «los ocho». La unión de rutas de T002, T014 y T017 también da ocho. Con «seis», `cierre.md` puede listar de menos en la capa 3 del informe final (los guiones de `testdata/` los revisa una persona) o la comprobación de FR-081 marcar dos guiones legítimos como ajenos. | Cambiar «seis» por «ocho» en las dos frases de T025 y nombrar los ocho ficheros. |
| B1 | Ambigüedad | BAJA | spec.md SC-005 (l. 314); plan.md l. 283-288; tasks.md T015 | SC-005 fija «≤ 3 800 bytes con cinco bloques» sin decir qué bloques. El plan mide 3 795 con `a21`-`a25` de la LPAC (5 bytes de holgura) y admite que con ids más largos (`a1-30` de la LCSP) se pasa. `TestLoQueLeeLaSkill` es determinista, así que no es flaky, pero cualquier cambio de redacción de la explicación de H7 FR 063 o de un campo de `data` lo rompe. | Nombrar en SC-005 los ids de la medida («`a21`-`a25` de `BOE-A-2015-10565`»). En T015, que el test cite ese margen en el mensaje de fallo. |
| C1 | Cobertura | BAJA | spec.md FR-021 (l. 218); tasks.md trazabilidad (l. 269) | FR-021 exige que las lecturas se ordenen por el orden de aplicación de sus entregas, también con invocaciones concurrentes. Ningún test de T006-T008 lo ejerce con concurrencia; `h7-grafo-concurrencia.txtar` solo pierde una aserción en T002 y no mira `lecturas`. Lo cubre, sin prueba propia, la serialización de `Apply` de H7 FR 014. | Un caso en `TestApplyConLecturas` (T007), o un párrafo en el plan que fije que FR-021 se hereda de H7 FR 014. |
| D1 | Terminología | BAJA | data-model.md §2 y tasks.md T006/T007 frente a T012 y `internal/graph/lectura.go:33` | «Lectura» significa dos cosas: `grafo.Lectura` (lo que vio una invocación de `boe articulo` de un bloque, T006) y `graph.Lectura` (el manejador de solo lectura de `world.db` que abre `Leer`, con `Lectura.Instantanea` en T012). Las dos conviven en T007, que toca `lectura.go` y `aplicar.go`. Es correcto en Go (paquetes distintos) pero confunde a quien lea una tarea. | En T006/T007, decir «lectura de un bloque» o `grafo.Lectura`, y «manejador `graph.Lectura`» en T012. No renombrar ningún tipo (H7 ya lo fija). |
| E1 | Cobertura / proporcionalidad | BAJA | tasks.md T005 y supuestos.md l. 11-13 | T005 y T003 retiran muchos tests y dejan el código que los cubría como regla genérica sin test propio (`validarOperacion`, `DatosCanonicos`, `indexar`…). Es coherente con FR-075 y con la constitución («Gates»), pero baja la cobertura de `internal/core/grafo`, que tiene umbral ≥ 85 %. T005 la mide, y T025 permite añadir tests si queda por debajo. | Sin cambios: el control ya existe. Al implementar, medir tras T005 y no después de T025. |
| F1 | Alineación con la constitución | BAJA (informativa) | tasks.md T002, T010, T014, T017, T018 | Cinco tareas `[datos]` modifican o crean ficheros de `testdata/` y `schemas/`. La constitución (capa 3) exige revisión humana de esos cambios tras el run. `plan.md` (capa 3) y T025 punto 1 los declaran, y ninguna graba nada de una fuente. | Sin cambios; cubierto una vez corregido A1. |

## Cobertura

Cada FR y SC con trabajo construible tiene al menos una tarea; los que dependen del cierre del workflow o de una persona lo declaran.

| Requisito | ¿Tarea? | Tareas | Notas |
|---|---|---|---|
| FR-001 a FR-007 (ámbito, argumentos, códigos, descripción) | Sí | T001, T012, T013, T014 | |
| FR-010 a FR-015 (cota, `data`, ≤ 40 000 bytes) | Sí | T001, T014, T015 | FR-013 se mide en T015 |
| FR-020, FR-022 (lecturas) | Sí | T001, T006, T007 | |
| FR-021 (orden con concurrencia) | Parcial | T006, T007 | Ver C1 |
| FR-023 a FR-026 (`version-obsoleta`, `world.db` de H7) | Sí | T001, T006-T008 | |
| FR-027 (orden inverso retirado) | Sí | T002 | |
| FR-030 a FR-032 (`fuente-caducada` solo sobre lo vigente) | Sí | T001, T009 | |
| FR-040 a FR-047 (`SKILL.md` v0.1.1) | Sí | T013, T014, T023 | |
| FR-048, FR-097 (CHANGELOG) | Sí | T024 | |
| FR-050 a FR-055 (formato de eval, `Juzgar`, informe, eval 19) | Sí | T018-T022 | La tasa de FR-055 la mide el job del cierre |
| FR-060 a FR-064 (salida legible) | Sí | T001, T016, T017 | |
| FR-070 a FR-073, FR-077 (regla genérica y retirada) | Sí | T001-T004 | |
| FR-074 a FR-076, FR-078 (`Persona`, lote, desempate) | Sí | T005 | |
| FR-080, FR-081 (guiones de H7) | Sí | T002, T014, T017, T025 | Ver A1 |
| FR-082 (esquemas) | Sí | T014, T018 | |
| FR-083 (sin ADR ni edición de H7) | Sí | T025 | |
| FR-090, FR-096 (`make ci`, cobertura) | Sí | todas, T005, T025 | |
| FR-091 a FR-095 | Sí | T001, T010, T011, T015, T019, T020, T023 | |
| FR-098 (job de evals) | Por diseño, no | cierre del workflow | No es una tarea |
| SC-001, SC-002, SC-005 | Sí | T015 | Ver B1 |
| SC-003, SC-004, SC-009, SC-010, SC-011 | Sí | T001, T008, T009, T014, T016, T017, T003, T004 | |
| SC-006 | Parcial | T021 y el job del cierre | La tasa la da el workflow |
| SC-007 | Fuera del run | persona, al leer el informe | Supuesto S2 |
| SC-008, SC-012, SC-013, SC-014 | Sí | T019, T020, T023, T008, T003-T005, T025 | |

## Alineación con la constitución

Sin violaciones.

- **III (tests primero):** T001 escribe la suite `[aceptacion]` antes de cualquier código y la verifica en rojo por una aserción.
- **IV (hexagonal):** reglas y `Comprobacion` en `internal/core/grafo` (solo biblioteca estándar); SQL en `internal/graph`; los tests que tocan SQL están en `internal/graph`. La medida se siembra con `Apply` (R3).
- **V (simplicidad):** ninguna dependencia nueva; cada mecanismo se traza en el plan; lo que solo servía a estados sin vía real se retira.
- **VII (grafo y privacidad):** la regla de `Persona` con NIF, NIE o DNI se queda con sus casos ASCII; las filas de `lecturas` solo nombran ids que ya guardan procedencia.
- **VIII (skills primero):** la eval 19 (T022) va antes que `SKILL.md` (T023); `SKILL.md` < 300 líneas y su tabla se regenera sin drift.
- **Gates, capa 1:** cada tarea declara sus rutas (verificado contra la regla «x.go declarado ⇒ x_test.go»); ningún `t.Skip`, `//nolint` ni TODO; las derivadas se producen con un programa fuera del repositorio y no hay grabación nueva.
- **Gates, capa 2 (ADR 0028):** hay umbral de materialidad (la regla genérica sustituye la enumeración de estados), criterio de uso (tabla «Uso, de fuera adentro» y `contracts/applet-graph.md` §3 y §6, con bytes medidos) y proporcionalidad (trazabilidad del plan).

## Comprobaciones contra el repositorio

- Ningún guion de `h7-*.txtar` fuera de `show` ejecuta `graph stats|show|check` sin `--json` y con éxito, así que la salida legible de T016 no rompe nada. Los guiones que sí las ejecutan sin `--json` esperan 2 o 3, y esos códigos no cambian.
- Ningún guion afirma `schema_version`, así que pasar el esquema a la versión 2 en T007 no exige tocar guiones.
- Ninguna `SKILL.md` tiene hoy un argumento opcional (`\[--`), de modo que el cambio del generador en T013 no produce drift.
- La descripción de `check` del contrato («lista como mucho 50 hallazgos») casa con la expresión `como\s+mucho\s+50\s+hallazgos` de la precondición de T001.
- Las cifras de T015 son consistentes: 270 normas + 2 160 bloques + 2 160 redacciones vistas = 4 590 `fuente-caducada`, más 240 `version-obsoleta` = 4 830, y `omitidos` = 4 830 − 50 = 4 780.

## Tareas sin requisito

Ninguna. T001, T002, T010, T011, T013, T017, T018 y T025 son de material, de control o de cierre, y cada una declara su razón en las excepciones de la cabecera de `tasks.md`.

## Métricas

- Requisitos: 66 FR y 14 SC (80)
- Tareas: 25 (T001-T025)
- Cobertura de requisitos con trabajo construible: 100 %. Tres requisitos son parciales o quedan fuera de la lista de tareas por diseño: FR-021 (C1), FR-098 y SC-007
- Ambigüedades: 1 (B1)
- Duplicaciones: 0
- Inconsistencias: 2 (A1, D1)
- Hallazgos CRÍTICOS: 0 · ALTOS: 0 · MEDIOS: 1 · BAJOS: 5

## Próximos pasos

- Sin CRÍTICOS: se puede continuar con `/speckit-implement`.
- Corregir A1 en `tasks.md` (T025: «seis» → «ocho», con los nombres de los ocho guiones).
- Opcional: B1 (nombrar los ids en SC-005), C1 (un caso de orden de lecturas concurrentes, o declarar en el plan que se hereda de H7 FR 014) y D1 (precisar «lectura de un bloque» en T006/T007).
