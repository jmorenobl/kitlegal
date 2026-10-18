// Package fiestas es el adaptador de la fuente boe.fiestas-laborales: la
// resolución de la Dirección General de Trabajo que publica cada año en el BOE
// la relación de fiestas laborales nacionales y de cada comunidad autónoma del
// año siguiente (art. 45.4 del Real Decreto 2001/1983), en el documento del
// diario https://www.boe.es/diario_boe/txt.php?id=<id> (fila
// boe.fiestas-laborales de docs/SOURCES.md; ADR 0039).
//
// No consulta la fuente en ejecución: ningún applet pide una resolución. Las
// tres de 2025, 2026 y 2027 las pide una sola vez el paso grabar_datos del
// workflow, sin modelo, con TestGrabarFiestasLaborales (grabacion_test.go,
// etiqueta grabacion) y el manifiesto testdata/grabaciones.json, por
// internal/httpx —solo GET, robots.txt, User-Agent identificable— y al ritmo de
// boe.legislacion-consolidada, que es el de su fila, y las deja grabadas en la
// raíz de evidencias que da KITLEGAL_EVIDENCIAS
// (evidencias/h9/boe.fiestas-laborales/). De lo grabado sale, por código y
// nunca a mano, data/festivos/nacionales-y-autonomicos.yaml (FR-020 de H9;
// contracts/datos-de-plazos.md §1.1, §1.3 y §4.1; research D14 y D15 de H9).
package fiestas
