---
id: H27
fase: "3"
---

# H27 · El Tribunal Constitucional por el BOE: `jurisprudencia` lee la sentencia sin que la persona la traiga (ADR 0040)

- **Objetivo**: que `jurisprudencia` resuelva y lea una sentencia del Tribunal Constitucional, y una del Supremo que anula una disposición, desde el BOE, sin pedir el documento a la persona. Todas las del TC se publican en el BOE, con reutilización autorizada, y en `ECLI:ES:TC:AAAA:N` la `N` es el número con que el BOE titula la sentencia (`docs/JURISPRUDENCIA.md`, §1, verificado el 2026-10-02).
- **Entrega**: la resolución de un ECLI, o de un número y año de sentencia, del Constitucional a su `BOE-A-…`, con su fecha, su número y su URL oficial, por la vía que fije el plan sobre `docs/JURISPRUDENCIA.md` (§1 y §4, puntos 1 y 2); `jurisprudencia` cita su texto como cita una norma, y sigue con el patrón de H23 para cualquier otro órgano. Emite al grafo la sentencia por su ECLI (ADR 0014).
- **Alcance**: solo lo que publica el BOE. El CENDOJ no cambia (ADR 0036): ni búsqueda automática, ni descarga. Una sentencia del TC que no se resuelve es la fila de «no encontrado» (ADR 0023) y no se cita.
- **Controles**: fixtures grabados del BOE; golden de la resolución de un ECLI que existe, de uno que no y de una sentencia del Supremo que anula una disposición; evals de `jurisprudencia` con una sentencia del TC que la persona no trae, juzgadas por su juez (ADR 0037).
- **Aceptación**: «¿qué dijo el Constitucional en la STC N/AAAA sobre X?» se responde con el texto de la sentencia leído del BOE y su cita, sin pedir el documento.
