# Dossier: DAG bloqueado

Causa mayor (ADR 0018): 3 tareas seguidas en cuarentena sin ninguna en verde entre medias.
Lo más probable es que las tareas pendientes dependan de lo que está en cuarentena.

## Tareas en cuarentena
- T001: 3 intentos sin quedar en verde (3 intentos; parche specs/023-h9-internal-core-plazos/gates/cuarentena/T001.patch)
- T002: 3 intentos sin quedar en verde (3 intentos; parche specs/023-h9-internal-core-plazos/gates/cuarentena/T002.patch)
- T003: 3 intentos sin quedar en verde (3 intentos; parche specs/023-h9-internal-core-plazos/gates/cuarentena/T003.patch)

## Tareas pendientes
- [ ] T001 [aceptacion] Los cinco guiones `testscript` de la entrega, escritos desde spec.md (US1.1, US2.1 a US2.3, US5.1 y US5.2; FR-001 a FR-013, FR-026, FR-028, FR-085) y desde contracts/applet-pla
- [ ] T002 [datos] Los tres manifiestos de grabación y sus tests `//go:build grabacion`, y **nada más**: ningún analizador, ninguna derivación, ningún fichero de `data/` (contracts/datos-de-plazos.md 
- [ ] T003 El dominio de los plazos, primera parte: los tipos, el calendario y la forma canónica de los ficheros de festivos, puros y sobre datos sintéticos, y **nada más** (data-model §1; contracts/d
- [ ] T004 El dominio de los plazos, segunda parte: el cómputo de la LPAC, del silencio y del contencioso, los avisos, los errores de argumentos y la salida, con la rejilla de las propiedades, todo so
- [ ] T005 [datos] Los tipos y la cobertura por materia en `data/`, con el esquema y el lector de la cobertura, la tercera raíz de grabaciones y la verificación de los preceptos contra lo grabado (con
- [ ] T006 [US3] Lo que `territorio` da a los dos adaptadores de los boletines: los municipios de una comunidad con sus formas, las comunidades con sus nombres y la forma compacta de un nombre, una so
- [ ] T007 [datos] [US3] Los festivos nacionales y autonómicos, derivados de las tres resoluciones del BOE grabadas tras T002 (contracts/datos-de-plazos.md §1.1 y §1.3; data-model §2; research D14, S1
- [ ] T008 [datos] [US3] Los festivos locales de la Comunidad de Madrid, derivados de las siete disposiciones del BOCM grabadas tras T002, con el cruce con `data/territorio/`, las correcciones en su o
- [ ] T009 [datos] [US1] El applet `plazos` con su verbo `calcular` en los dos registros, con su esquema, sus golden, la garantía sin red y las listas literales que su llegada cambia, en una sola tare
- [ ] T010 [US1] La tabla de casos calculados a mano y el festivo de una parte, sobre el applet y los ficheros empotrados (contracts/casos-de-plazos.md; research D23, D30, S1, S3; plan.md paso 8). **`
- [ ] T011 [datos] [US6] El formato de eval gana `plazo`, `no_cubierto` y el comando `calcular`, con su juicio sin modelo, las reglas del conjunto de `legal-core` y las cinco evals, en una sola tarea 
- [ ] T012 [US6] Los umbrales `plazo_sin_fecha:<modelo>:<modo>` y `cobertura_sin_declarar:<modelo>:<modo>` en el informe del job, con `sin_activar` de `legal-core`, y los tests que fijan los umbrales 
- [ ] T013 [US4] La skill `legal-core` v1 y sus dos referencias generadas (contracts/skill-legal-core.md §1 a §4; contracts/SKILL-legal-core.prototipo.md; research D21, V13, V14, V22, M4; plan.md paso
- [ ] T014 Los mutantes de FR-089, aplicados uno cada vez en una copia de la cabeza fuera del árbol, y su registro (quickstart.md §11 y §12; plan.md paso 12 y fila de FR-089 de «Controles de umbral»).
- [ ] T015 Lo que los documentos dicen de `plazos`, de los festivos, de `legal-core`, de sus evals y de sus umbrales, con cada cifra, fichero, test y umbral sacados del árbol y no de memoria (research
- [ ] T016 Cierre de la Definition of Done, sin código de producto (`ROADMAP.md` §1; quickstart.md §0 a §13 y §16; plan.md, «Obligaciones para tasks.md»): specs/023-h9-internal-core-plazos/cierre.md c

## Qué hace falta para seguir
Leer la nota gates/tarea-Tnnn.md de cada tarea en cuarentena. Si la causa está en el spec o en el roadmap,
se corrige ahí y se relanza el hito. Si es del entorno (herramienta, red, permisos), se arregla y se reanuda
con scripts/hito.sh --resume <run_id>. Al reanudar, el contador de cuarentenas seguidas vuelve a cero; las
tareas que sigan en gates/cuarentena.json no se reintentan: para reintentar una, se borra su entrada.
