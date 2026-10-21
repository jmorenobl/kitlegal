---
id: H34
fase: "1"
---

# H34 · El puerto de fuentes genérico, con el BOE y EUR-Lex

- **Objetivo**: que una fuente nueva se conecte con el mismo contrato que el BOE, y que la normativa europea se lea y se cite como la española.
- **Entrega**: el contrato de una fuente por capacidades —buscar, leer, versiones, metadatos—; un registro de fuentes en `data/` con sus capacidades, condiciones y cobertura, coherente con `docs/SOURCES.md`; el BOE sobre el contrato nuevo sin cambiar su salida; EUR-Lex como segunda implementación, tras verificar sus condiciones de reutilización y la vía de un programa; y un ADR del conector de terceros —bases de jurisprudencia o doctrina de pago con API—: credenciales de la persona dentro del producto, licencias y lo que se envía al tercero, sin construir ninguno.
- **Riesgos**: las condiciones y los formatos de EUR-Lex; un segundo identificador natural (CELEX, ELI europeo) en la caché y el grafo.
- **Controles**: los de una fuente nueva (fixtures, golden, `verify-sources`) y las evals de una pregunta de normativa europea.
