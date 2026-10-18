// Package bocm es el adaptador de la fuente bocm.fiestas-locales: la
// resolución de la Dirección General de Trabajo de la Comunidad de Madrid que
// declara cada año, en el Boletín Oficial de la Comunidad de Madrid, las
// fiestas laborales de ámbito local de cada municipio (art. 46 del Real
// Decreto 2001/1983), y cada resolución que la modifica, en XML en
// https://www.bocm.es/boletin/CM_Orden_BOCM/AAAA/MM/DD/<id>.xml, con la fecha
// de publicación que lleva el identificador BOCM-AAAAMMDD-N (fila
// bocm.fiestas-locales de docs/SOURCES.md; ADR 0039).
//
// No consulta la fuente en ejecución: ningún applet pide una disposición, y el
// buscador del BOCM está cerrado a los agentes, así que la lista de
// disposiciones de cada año la fija una persona en el manifiesto antes del run.
// Las siete de 2025 y 2026 las pide una sola vez el paso grabar_datos del
// workflow, sin modelo, con TestGrabarFiestasLocales (grabacion_test.go,
// etiqueta grabacion) y el manifiesto testdata/grabaciones.json, por
// internal/httpx —solo GET, robots.txt, User-Agent identificable— y a 10 s
// entre peticiones, el Crawl-delay de su robots.txt, y las deja grabadas en la
// raíz de evidencias que da KITLEGAL_EVIDENCIAS
// (evidencias/h9/bocm.fiestas-locales/). De lo grabado sale, por código y nunca
// a mano, data/festivos/locales/13.yaml (FR-020 de H9;
// contracts/datos-de-plazos.md §1.2, §1.3 y §4.1; research D13 y D15 de H9).
package bocm
