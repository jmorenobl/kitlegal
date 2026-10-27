# Contrato: el apartado de `docs/SKILLS.md` §1 (FR-014) y la entrada de `CHANGELOG.md` (FR-016)

## 1. `docs/SKILLS.md` §1

Un párrafo nuevo en §1, «Qué va en el cuerpo y qué en una referencia», detrás del que dice de qué dos maneras el cuerpo
dice cuándo leer cada referencia (el que termina con el ejemplo de `buscar_la_norma`, líneas 62 a 71) y delante de
«Una referencia es de una de dos clases». Nada más de §1 ni del resto del fichero cambia.

> **Lo que se lee y lo que no.** Lo midió el sondeo de la eval 15 de `boe-legislacion` (`evidencias/sondeos-20261025/`):
> las sesiones abren las referencias de las que necesitan un dato para seguir y no las que solo traen guía, aunque el
> cuerpo mande leerlas. Ninguna de las 40 sesiones de `brazo-D` abrió la que el cuerpo mandaba leer antes de
> responder; en los seis brazos se abrieron `normas`, de 23 a 40 sesiones por brazo, y, en los brazos con la skill de
> H31, `con_herramientas`, de 6 a 9. De ahí dos consecuencias: lo que toda respuesta necesita va en el cuerpo, y una
> referencia nueva lleva el paso del protocolo que la pide, con la condición de su clase. Lo respaldan tres prácticas
> de las guías de autoría de la especificación ([agentskills.io](https://agentskills.io/)) y de Anthropic —sin volver a
> consultarlas: [Las dos fuentes](#las-dos-fuentes)—: decir cuándo leer cada referencia, no solo que existe; explicar
> el porqué en vez de encadenar prohibiciones; y quitar del cuerpo lo que el agente no haría mal sin ello.

- **Las cifras** son las de `evidencias/sondeos-20261025/LEEME.md`, líneas 57 a 59 («Ninguna de las 40 sesiones de D
  abrió `responder.md`»; `normas`, «de 23 a 40 sesiones por brazo»; `con_herramientas`, «de 6 a 9»), y `brazo-D` es
  el que llevaba la forma y los ejemplos en `references/responder.md`, que el cuerpo mandaba leer antes de responder
  (línea 45). Las coteja la revisión final (SC-013).
- **Las URL**: la única que el párrafo añade es `https://agentskills.io/`, que registra `web/src/components/Pie.astro`
  (línea 30) en `main`. La de la guía de Anthropic no la registra ningún fichero del repositorio: el párrafo enlaza a
  «Las dos fuentes» (`#las-dos-fuentes`, el ancla del índice de la línea 14) y el informe final la da como pendiente
  de la persona.
- **Las tres prácticas** son las que la entrada del hito atribuye a las dos guías; el párrafo no las cita
  literalmente, porque el run no tiene red para consultarlas (spec, «Assumptions»).

## 2. `CHANGELOG.md`

Una entrada en *Unreleased*, la primera de «### Cambiado» (que va de la más reciente a la más antigua), en la forma de
las de H31. La de v0.3 no se toca.

> - **`boe-legislacion` v0.3.1: la respuesta empieza por lo que se pregunta, lo que nombra se lee antes y se repasa
>   antes de enviarla** (H31.1). Sustituye a `boe-legislacion` v0.3. Lleva, palabra por palabra, los pasos que se
>   midieron en el sondeo de la eval 15 (`evidencias/sondeos-20261026/`, `brazo-W`: 1 de 40 en `afirma_lo_no_leido`
>   y 2 de 40 en `cuenta_su_proceso`): la primera viñeta del paso 4 dice que la respuesta es todo lo que se escribe
>   tras la última orden o llamada, que quien pregunta la lee entera y que habla de la norma, no de la consulta; el
>   paso 3 lee antes de escribir el precepto al que remite un bloque leído, o cuyo id solo salió en un índice, si la
>   respuesta va a hablar de él, y lo que remita esa segunda lectura ya no se lee; el paso 5 no junta número y materia
>   de lo no leído; y el paso 6, nuevo, repasa la respuesta frase a frase antes de enviarla. Para que el cuerpo siga en
>   11 000 caracteres como mucho —queda en 10 995—, lo demás se dice más corto, sin quitar ninguna regla invariante ni
>   forma fija, y la regla de `kitlegal graph check` que no termina con `0` queda en `lecturas_que_fallan`, como la
>   frase de «no encontrada con esta búsqueda» en `buscar_la_norma` y lo que no lleva la respuesta de la comprobación
>   en `redaccion_modificada`. Las evals 14, 15 y 17 llevan en `lecturas` los nueve preceptos que sus respuestas
>   nombran, grabados para que las sesiones los lean sin red.

Cada afirmación de la entrada sale del diff de la entrega con `main`: los pasos, de §1 de
[skill-boe-legislacion-v0-3-1.md](skill-boe-legislacion-v0-3-1.md); el cuerpo de 10 995 caracteres, del mismo
contrato; la regla 7 de `main` («Si `kitlegal graph check` no termina con `0`»), que no está en la skill medida, y las
dos referencias que cambió la #200, del diff de `buscar_la_norma.md` y `redaccion_modificada.md` con `main` (research
V2); las lecturas, de [lecturas-y-grabaciones.md](lecturas-y-grabaciones.md) §1.
