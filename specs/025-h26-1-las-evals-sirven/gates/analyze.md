# Informe de análisis · H26.1 (spec.md, plan.md, tasks.md)

Sesión de un paso, solo lectura salvo este fichero. **No se ha ejecutado ningún test ni `make ci`**: lo que sigue sale de
leer los tres artefactos, la constitución citada en ellos y el árbol (`grep`, `sed`, `ls`, `wc`). Lo que no pude medir
va marcado como tal. El hook opcional `before_analyze` (`speckit.git.commit`) no se ejecutó: la sesión es de solo lectura.

## Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| F1 | Inconsistencia | LOW | spec.md «Criterios del hito» (l. 542-546) frente a SC-014 (l. 497-499) | La entrada del hito, transcrita, enumera cinco umbrales que se cumplen; SC-014 enumera seis, con `lectura_sin_intentar` (FR-033, añadido en clarify). | Ninguna. La diferencia es deliberada, y SC-013 y FR-033 la cubren. |
| F2 | Subespecificación | LOW | FR-025, T013, T015 | «Un test que ate» el prefijo del arnés con el mensaje del binario: `TestLineaDeLaSalidaNoEscrita` queda en verde con el mutante por construcción (el prefijo sale de `render.Nuevo`). Lo que se pone en rojo es el literal de `TestTuberiaCerrada`, que ya existe. | Ninguna. El criterio de FR-025 (el mutante hace fallar algún test) se cumple y T015 lo registra. |
| F3 | Alcance de tarea | LOW | T004 | Es la tarea más ancha: tres ficheros de producción, cinco tests y dos mutantes apuntados. Sigue siendo una rebanada vertical, pero cualquier rojo cuesta un intento entero. | Ninguna. Se puede partir en la implementación si el guardián o el tiempo lo piden. |
| F4 | No medido | LOW | T007, T008, T010, T014 | Las cifras de umbrales (14 y 12 tras T008; 18 y 12 en `legal-core` y 19 y 11 en las otras dos tras T010) vienen de research y de los contratos. No las he vuelto a contar contra `umbrales.go`. | Las miden los propios tests (`TestUmbralesDeLasTresSkills`); si el árbol contradice un contrato, manda el árbol y se anota en `supuestos.md`. |

Sin hallazgos CRITICAL ni HIGH.

## Comprobado contra el árbol en esta sesión

- Los lectores y funciones que las tareas nombran existen: `ConsultasNecesarias` (`consultas.go:105`),
  `(*consultasNecesarias).anotarBloque` (`:178`), `prepararLaCache` (`preparar.go:319`), `queFaltaDesde` (`:105`),
  `leerLasEvalsDeHoy` (`medida.go:1056`), `satisface` y `consultoComoPide` (`juzgar.go:1019` y `:1057`),
  `informarLaJuzgada` (`:968`), `lecturaDeLosArgumentos` (`:1112`), `esDeLaNorma` (`:1150`), `contieneComoPalabra`
  (`:1184`), y `lecturasDeLosTipos` y `lecturasDeLaCobertura` (`grabaciones_test.go:1037` y `:1084`).
- `render.Nuevo` y `enDescriptor` (`render.go`, `"render: no se pudo escribir en %s: %w"`) existen. `depguard` solo prohíbe
  `internal/render` a `internal/core/**` y a `internal/graph`, así que `internal/evals` puede importarlo. Hoy `internal/evals`
  no lo importa: lo introduce T013.
- `skills/legal-core/SKILL.md` tiene 299 líneas, el máximo; T011 compensa lo que añade.
- El peor caso de `legal-core` cuadra: hoy `485 + (⌈55/2⌉ + ⌈54/2⌉ + ⌈6/2⌉) × 272 = 16261 s`; con 12 sesiones más por modo
  (2 evals × 3 repeticiones × 2 modelos), `⌈67/2⌉ + ⌈66/2⌉ + ⌈6/2⌉ = 70` y `485 + 70 × 272 = 19525 s`, que cabe en los
  352 min (21 120 s) y deja fuera del caso de 288 min (17 280 s).
- La eval 05 trae lo que FR-017 y T008 copian: los tres comandos, `a124` con `fecha: "2026-09-09"`, `plazo.fecha`
  `"2026-10-13"` y `razonamiento`. `daprimera` existe en `data/cobertura.yaml` para `BOE-A-2015-10565`. Hoy hay diez evals en
  `evals/legal-core/`, así que con la 11 y la 12 son doce, como dice T008.
- No existe ningún esquema del informe fuera de Go que obligue a una tarea `[datos]` para los campos nuevos
  (`otras_fallidas` y `fuera_de_lo_grabado` solo aparecen en Go y en los gates de hitos pasados).
- Cada tarea declara en «Rutas:» los ficheros que toca. Los `x_test.go` de cada `x.go` declarado cubren los tests
  nombrados, y `mutantes.md` está en T004, T007, T008, T013 y T015.

## Cobertura de requisitos

| Clave | ¿Tarea? | Tareas | Notas |
|-------|---------|--------|-------|
| FR-001 a FR-005 | sí | T004 (T015) | derivación de `data/`, falta si no está grabado y base de la medida |
| FR-006, FR-007 | sí | T009 (T006 para FR-007) | `lecturas` sin exigirse |
| FR-008, FR-009, FR-012, FR-014 | sí | T001, T003 | esquema y rechazos de `LeerEval` |
| FR-010, FR-011 | sí | T002, T009 | la grabación la hace `grabar_datos` |
| FR-013 | sí | T005, T008 (T015) | `sin_servir` solo en las sesiones de su eval |
| FR-015, FR-016 | sí | T006, T007, T008, T013 | reparto, serie y comando que admite 4 y 5 |
| FR-017 a FR-020 | sí | T006, T008 | evals 11 y 12, umbrales existentes y tope |
| FR-021, FR-022 | sí | T006, T010 | `fuera_de_lo_grabado`, `decide: false` |
| FR-023 a FR-028 | sí | T012, T013 (T015 para FR-025) | `sin_sobre` y `salida_no_escrita` |
| FR-029 a FR-031 | sí | T011 | `legal-core` v1.2.3, 299 líneas |
| FR-032 | sí | T014 | CHANGELOG y CONTRIBUTING |
| FR-033 | sí | T007, T008, T013 (T015) | `lectura_sin_intentar` decide |
| SC-001 a SC-010, SC-015, SC-018, SC-021 | sí | ver «Requisitos → tareas» de tasks.md | los fija `make ci` |
| SC-011 a SC-014, SC-016, SC-017, SC-019, SC-020 | sí, indirecto | T004, T007, T008, T010, T013, T016 | los mide el job de cierre; en `make ci` los fijan los tests de la tabla |

## Alineación con la constitución

Sin conflictos. Todo cuadra con los principios que rigen el hito:

- Sin telemetría ni red nueva: ninguna tarea toca la red, y la grabación es del paso `grabar_datos`.
- El binario no analiza respuestas: el motivo `salida_no_escrita` se reconoce con la línea de error del propio binario, no
  con texto del modelo.
- Ningún umbral se cumple cambiándolo: `fuera_de_lo_grabado` se publica sin decidir y `lectura_sin_intentar` decide.
- Aceptación e2e «no aplica», justificada: `cmd/kitlegal` no enlaza `internal/evals`.
- Definition of Done mapeada en tasks.md.

## Tareas sin requisito

Ninguna. T001 y T002 (`[datos]`), T014, T015 y T016 (pulido y cierre) están mapeadas a FR y SC.

## Métricas

- Requisitos: 33 FR y 21 SC.
- Tareas: 16 (2 `[datos]`, 0 `[aceptacion]`).
- Cobertura de requisitos con al menos una tarea: 100 %.
- Ambigüedades: 0. Duplicaciones: 0. Incidencias CRITICAL: 0 (HIGH: 0, MEDIUM: 0, LOW: 4).

## Siguientes pasos

Se puede pasar a implementar. No hay nada que resolver antes. F1 a F4 no cambian el orden ni el contenido de las tareas.
