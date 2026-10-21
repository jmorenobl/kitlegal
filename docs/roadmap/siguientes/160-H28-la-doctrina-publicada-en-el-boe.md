---
id: H28
fase: "2"
---

# H28 · La doctrina publicada en el BOE + skill `doctrina` (ADR 0040)

- **Objetivo**: que un despacho cite un dictamen del Consejo de Estado, una circular, instrucción o consulta de la Fiscalía General del Estado o un informe de la Abogacía del Estado con su texto delante, y que ninguna skill caracterice una doctrina que no ha leído. Las tres bases están en el buscador del BOE (comprobado el 2026-10-19; los dictámenes, desde 1987 y con el texto completo).
- **Entrega**: la primera tarea verifica, sin código de producto, la vía de un programa a cada base (`robots.txt`, API de datos abiertos, formato, condiciones de reutilización) y su fila en `docs/SOURCES.md`. Si un programa puede leer un documento por su identificador, un applet lo trae con el sobre de siempre; si además puede buscar, busca. Si no puede buscar, el hito sigue el patrón de `jurisprudencia` (ADR 0036): la skill prepara la búsqueda, la persona busca y trae, y el binario resuelve el identificador y coteja. Skill genérica `doctrina`, con las reglas de `jurisprudencia`: nada citado sin leerlo, y la doctrina se distingue de la norma y de la sentencia.
- **Alcance**: fuera, la doctrina que no publica el BOE —DGT y TEAC con su vertical, tras revisar sus términos; AEPD, TACRC y Consejo de Transparencia, desde el backlog de fuentes— y la doctrina académica.
- **Controles**: los de una fuente nueva (fixtures, golden, `verify-sources`); evals de `doctrina` y su juez, medido contra respuestas reales antes de decidir (ADR 0037).
- **Aceptación**: «¿qué ha dicho el Consejo de Estado sobre X en el dictamen N/AAAA?» se responde con el texto del dictamen y su cita, o con la búsqueda preparada y el cotejo del documento que trae la persona.
