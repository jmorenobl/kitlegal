## Specification Analysis Report

**Hito**: H31.1 · `boe-legislacion` habla de la norma, no de la consulta, y lee lo que nombra
(`specs/028-h31-1-boe-legislacion-habla/`). Análisis de solo lectura de `spec.md` (448 líneas), `plan.md` (311) y
`tasks.md` (235); `research.md`, `data-model.md`, `contracts/` y `quickstart.md` se han consultado solo por las cifras y
los textos que citan las tareas. La constitución no se ha releído entera en esta sesión: la alineación se juzga por lo que
el plan declara en su «Constitution Check» y por las decisiones cerradas de `CLAUDE.md`.

### Hallazgos

| ID | Categoría | Severidad | Ubicación | Resumen | Recomendación |
|----|-----------|-----------|-----------|---------|---------------|
| F1 | Underspecification | LOW | spec.md:L40-41, L257-258; tasks.md:L94 | El cuerpo de v0.3.1 queda en 10 993 caracteres de un máximo de 11 000: 7 de margen. Cualquier retoque del texto del contrato o de la tabla de comandos regenerada lo pone en rojo (`make skills-sync`). El spec ya prevé qué hacer (Edge Cases: sacar lo que solo pide una clase de pregunta), pero T003 no dice qué sale primero si la cifra real difiere de la del contrato. | Sin cambio obligatorio. Si T003 se pasa de 11 000, aplicar la regla de FR-006 (otro texto de una sola clase de pregunta a una referencia a mano, con su enlace) y no recortar FR-001 a FR-004. |
| F2 | Inconsistency | LOW | tasks.md:L76, L94, L75 | Las líneas de tarea citan rutas que la tarea dice no tocar (T002: `.github/workflows/evals.yml`, `internal/evals/definicion.go`; T003: `docs/SKILLS.md`). Según el extractor de rutas del guardián (memoria «etiqueta literal en texto de tarea»), toda ruta que aparece en la línea se declara y por tanto queda permitida: el alcance de la tarea se ensancha en la práctica. No lo prueba ningún test de esta sesión. | Redactar esas menciones sin carpeta («el workflow de evals», «`definicion.go`», «la guía») o aceptar el ensanchamiento, que no rompe nada. |
| F3 | Coverage / trazabilidad | LOW | spec.md (sin mención); plan.md:L52, L125, L250, L301; tasks.md:L94 | `CONTRIBUTING.md` (la frase de «la regla 7 sin su punto final», línea 743) se toca en T003 y no aparece en ningún FR ni en «Fuera de alcance» del spec; solo el plan lo justifica como consecuencia de FR-006 (sale la regla 7). Verificado: `grep -c 'regla 7 sin su punto final' CONTRIBUTING.md` da hoy 1, y T003 espera 0. | Aceptable tal cual: la tabla de trazabilidad de tasks.md lo asigna a FR-006. Si se quiere cerrar, añadir una línea de «Alcance» al spec en una vuelta posterior; no bloquea. |
| F4 | Ambiguity | LOW | spec.md:L224 | FR-001: «Una regla que el cuerpo de hoy ya dice con el mismo contenido se queda como está» depende de un juicio de identidad de contenido. Lo acota la tabla del plan (contracts/skill-boe-legislacion-v0-3-1.md §2), que la revisión final coteja (SC-003). | Sin cambio; la revisión final debe cotejar la tabla fila a fila. |
| F5 | Coverage (por diseño) | LOW | spec.md SC-007, SC-009–SC-012; tasks.md:L115-121, L183 | SC-007, SC-009, SC-010, SC-011 los mide el job de evals del cierre y SC-012 un sondeo de una persona: ninguna tarea los construye ni este análisis puede medirlos. Las tareas solo entregan el texto que se mide. | Sin acción; el informe final debe dar SC-012 como pendiente (ya lo dice FR-013/SC-012). |
| F6 | Consistency note | LOW | spec.md:L438-440; tasks.md:L76 | «Fuera de alcance» excluye cambios en el arnés de evals «salvo el tope … y lo que sus tests esperan»; T002 cambia `internal/evals/conjunto_test.go` y `preparar_test.go` (los dos tests que enumeran `lecturas`). FR-011 y SC-004 los exigen, así que no hay conflicto, pero la excepción del spec solo menciona el tope. | Sin cambio; la lectura coherente es que los tests de `lecturas` de H26.1 son parte de FR-011. |

Sin hallazgos CRITICAL ni HIGH. Sin duplicados entre requisitos. Sin marcadores sin resolver (`TODO`, `???`,
`<placeholder>`) en spec, plan ni tareas.

### Coverage Summary

| Requisito | ¿Tarea? | Tareas | Notas |
|-----------|---------|--------|-------|
| FR-001 (reglas de `27aeccc`, seis ejemplos) | Sí | T003 | tabla de reglas en contracts §2; la coteja la revisión final |
| FR-002, FR-003, FR-004 (leer lo que se nombra, un salto, genérico) | Sí | T003 | `sin-instrucciones-de-evals` y el grep de la tarea |
| FR-005, FR-006, FR-007, FR-008 | Sí | T003 | gates de H31 sin cambios |
| FR-009 (nueve `lecturas`) | Sí | T002 | test del mapa en rojo antes de las evals |
| FR-010 (manifiesto y grabación) | Sí | T001 + paso `grabar_datos` | única tarea con la etiqueta de datos entre las líneas de tarea |
| FR-011 (caché sirve los bloques) | Sí | T002 | `TestCacheDeLasSesiones`, `TestLecturaDeLaEvalSinGrabar` |
| FR-012 (guía §1) | Sí | T004 | |
| FR-013 (peor caso y tope) | Sí | T002 (ejecuta el test; sin cambio de tope) | ver comprobaciones |
| FR-014 (CHANGELOG) | Sí | T004 | cifra del cuerpo tras T003 |
| SC-001–SC-006, SC-013 | Sí | T001–T004 | |
| SC-007, SC-009, SC-010, SC-011 | Indirecto | T003 entrega; mide el job del cierre | por diseño |
| SC-012 | No | — | sondeo de una persona, fuera del run (ADR 0032) |

### Constitution Alignment Issues

Ninguno detectado. Verificado contra el texto de `CLAUDE.md` (decisiones cerradas): skills primero (el hito protege una
skill existente), genericidad por materia (FR-004 prohíbe nombrar las normas de las evals 14, 15 y 17 en las líneas
añadidas), sin cambios en el binario ni en umbrales, la fusión y el sondeo quedan en manos de una persona, y ninguna tarea
toca `scripts/workflow/`, `.specify/` ni `internal/workflow/`. La lectura íntegra de la constitución no se ha repetido en
esta sesión.

### Unmapped Tasks

Ninguna. T001→FR-010; T002→FR-009, FR-011, FR-013; T003→FR-001 a FR-008; T004→FR-012, FR-014.

### Comprobaciones hechas en esta sesión (ejecutadas y vistas terminar)

- `.specify/scripts/bash/check-prerequisites.sh --json --require-spec --require-tasks --include-tasks` → `FEATURE_DIR`
  correcto, con `tasks.md` presente.
- Cuerpo actual de `skills/boe-legislacion/SKILL.md` (tras el segundo `---`): **10 992** caracteres (`wc -m`), como dice
  el spec.
- Los ocho ids de FR-010 existen en los índices grabados con el título esperado: `a7` «Artículo 7», `a19`, `a20`,
  `dadecimoctava` «Disposición adicional decimoctava» de `BOE-A-2006-20764`; `a60`, `a92`, `a100`, `a104` de
  `BOE-A-2004-4214`.
- `internal/evals/testdata/grabaciones.json` ya tiene `a18` (Ley 35/2006), `a15` (RDL 2/2004) y `a24` (Ley 19/2013), como
  suponen el spec y T001; `evals/boe-legislacion/14` y `15` ya llevan `lecturas:` con una entrada cada una y la `17` no
  lleva la clave, como dice T002.
- `go test ./internal/evals -run TestDefinicionDelJob -count=1` → verde (151 pruebas del paquete). El `timeout-minutes`
  del trabajo de `boe-legislacion` en `.github/workflows/evals.yml:276` es 352 min = 21 120 s, y el plan da el peor caso
  en 20 117 s (research V3), coherente con FR-013.
- Fisher unilateral 2 de 40 frente a 16 de 80, recalculado con una hipergeométrica: p = 0,0234, como dice SC-012 (≈ 0,023).
- Entre las líneas de tarea de `tasks.md` solo T001 contiene la etiqueta de datos; ninguna contiene la de plataforma.

### No medido

- El cuerpo de 10 993 caracteres que promete el contrato y que los gates de contenido pasan sobre el texto nuevo:
  requiere implementar T003.
- `make ci` completo, los gates sobre la skill nueva y las lecturas sin red de las tres evals: no se ha ejecutado.
- `afirma_lo_no_leido` y `cuenta_su_proceso` con la skill nueva: dependen del job de cierre y del sondeo de una persona.
- La tabla de reglas de `27aeccc` frente al contrato (FR-001, SC-003): no se ha recotejado fila a fila; es trabajo de la
  revisión final.

### Metrics

- Requisitos funcionales: 14 · criterios de éxito: 13 · tareas: 4
- Cobertura (FR con ≥ 1 tarea): 14/14 = 100 %; SC con tarea o con medida declarada fuera del run: 13/13
- Ambigüedades: 1 (F4) · Duplicados: 0 · Problemas críticos: 0 · HIGH: 0 · MEDIUM: 0 · LOW: 6

### Next Actions

- Sin CRITICAL ni HIGH: se puede pasar a `/speckit-implement`.
- Opcionales, sin bloquear: F2 (redactar las rutas citadas de T002/T003 sin carpeta si se quiere un guardián estricto) y
  F1 (tener a mano el criterio de qué texto sale primero si T003 se pasa de 11 000).
- Lo que queda a una persona: el sondeo de SC-012 antes de fusionar.
