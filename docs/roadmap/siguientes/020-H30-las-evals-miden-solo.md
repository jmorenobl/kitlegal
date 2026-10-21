---
id: H30
fase: "1"
depende_de: [H29]
---

# H30 · Las evals miden solo lo que cambia: dependencias de cada eval y resultados por huella

- **Objetivo**: que una propuesta de cambio mida las evals que ha podido cambiar, y nada más. Con una skill de civil y otra de fiscal, tocar fiscal no mide civil; tocar las dos, sí. Hoy cualquier cambio en lo que las evals miden lanza las tres skills enteras en los dos modos.
- **Entrega**: (1) el conjunto de dependencias de cada eval, calculado y no declarado a mano: la skill que mide y la referencia de su materia, los `data/*.yaml` de los que se generan, los paquetes Go de los applets que invoca (de sus `comandos:` y `go list -deps`), sus grabaciones y, en una skill con juez, su rúbrica y sus casos; lo común —kernel, `httpx`, caché, arnés y job, modelo y versión de Claude Code— las dispara todas, y una ruta que no se sabe asignar también; (2) la descripción de una skill compite con las demás por la activación, así que tocarla dispara las evals de activación de todas; (3) el job construye su matriz con las evals elegidas; (4) el resultado de una eval se guarda con su huella —el modelo que decide, la del juez y la de sus dependencias— y una eval con la misma huella que una medida de `main` reutiliza su resultado, sin abrir sesión; (5) la suite completa antes de cada release y a petición; (6) un ADR decide cómo cuenta lo reutilizado en el umbral en tasa (ADR 0038) y en la serie de cierres.
- **Alcance**: no cambia qué mide ninguna eval ni sus umbrales, ni abre más sesiones a la vez contra la suscripción.
- **Controles**: un test en `make ci` que exige un conjunto de dependencias calculable para cada eval; mutantes: tocar la referencia de una skill elige sus evals y no las de otra, tocar `internal/cli` las elige todas, tocar una descripción elige la activación de todas; golden de la matriz.
- **Aceptación**: una propuesta que solo toca el cuerpo de `skills/jurisprudencia/SKILL.md` no abre ninguna sesión de `boe-legislacion` ni de `legal-core`.
