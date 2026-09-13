# 0012 · Base antes que vertical: la primera skill es genérica y las verticales la especializan

- **Estado**: aceptada
- **Fecha**: 2026-09-12
- **Hito**: transversal (cambia el alcance de H5 y H24 y añade la fase 8; la numeración no cambia)

> Nota (2026-09-13): desde el ADR 0013 las verticales viven en el backlog sin número (grupo
> «verticales»), no en una «fase 8»; H24 pasa al backlog («profundidad») y el mecanismo de packs va con la
> distribución, no con H22. El principio —base antes que vertical, verticales sin código y en cualquier
> momento tras H5— no cambia.

## Contexto y problema

El roadmap heredó de los documentos semilla que la única skill existente, `boe-fiscal` (Python, fuera de
este repo), fuera también la primera del producto: H5 la migraba «idéntica salvo `scripts/boe.py` →
`scripts/boe`», `data/normas.yaml` nacía con `vertical: fiscal`, las diez evals eran fiscales y la skill
genérica de consulta al BOE, `boe-legislacion`, no llegaba hasta H24 (fase 4), donde `boe-fiscal`
«pasaba a ser una especialización» suya.

Eso invierte el orden natural. La herramienta (applet `boe`, H4) es genérica desde el primer día: consulta
la API de Legislación Consolidada sin saber qué es un impuesto. Pero entre H5 y H24 la única skill que
sabía usarla se activaba por una descripción fiscal, de modo que una consulta sobre la LPAC, la LCSP o la
LRBRL, que son las que necesitan las skills municipales de las fases 1 a 3, no tenía skill propia. Y el
propósito del proyecto es que cualquier ciudadano o profesional disponga de una capa determinista para
toda la normativa, con las especializaciones por materia llegando después y poco a poco.

## Opciones consideradas

1. **Dejar H5 como estaba** y adelantar `boe-legislacion` a un hito nuevo entre H5 y H6. Obliga a
   renumerar todo el roadmap (ya renumerado una vez, ADR 0008) y construye primero la especialización y
   luego la base de la que se especializa.
2. **H5 entrega las dos skills**, `boe-legislacion` y `boe-fiscal`, a la vez. Hace más grande el hito que
   además estrena el andamiaje (`skills-sync`, evals, comprobaciones en `make ci`), justo el que conviene
   que sea pequeño.
3. **H5 entrega solo `boe-legislacion`** con el andamiaje, y la migración de `boe-fiscal` se convierte en
   la primera de una fase de verticales bajo demanda, que puede colocarse en cualquier punto tras H5
   porque una vertical es `SKILL.md`, datos y evals, sin código.

## Decisión

Se adopta la **opción 3**:

- Las skills nacen genéricas. La base la forman `boe-legislacion` (H5), `legal-core` (H7 y H9),
  `cita-verificada` (H8) y las skills municipales de las fases 2 y 3.
- Una **vertical** (fiscal, laboral, mercantil…) es una especialización de una skill base: reutiliza su
  protocolo y añade criterio, no herramientas. Se compone de normas con `vertical: <nombre>` en
  `data/normas.yaml`, `references/` generadas propias, calendario y reglas de la materia, evals, y un
  `packs/<vertical>/pack.yaml` que extiende `legal` cuando exista el mecanismo de packs (H22). Si una
  vertical necesita una fuente propia (PETETE, DYCTEA, AEAT…), la fuente entra por la fase 7 con sus TOS
  revisados.
- Las verticales viven en la **fase 8** del roadmap, un hito por vertical, bajo demanda y en cualquier
  momento tras H5. La primera es `boe-fiscal`, migración de la skill Python existente sobre
  `boe-legislacion`, con evals que reproducen sus consultas actuales para no perder comportamiento.
- H5 pasa a entregar `boe-legislacion`: su protocolo es el de `boe-fiscal` generalizado a cualquier
  materia; `data/normas.yaml` nace sin campo `vertical` (entra con la primera vertical) y con las leyes
  vertebrales que la skill referencia; las diez evals cubren materias distintas y al menos una reproduce
  una consulta que hoy resuelve `boe-fiscal`. H24 pasa a ser `boe-legislacion` v1 (vigilancia y
  novedades). H4 no cambia: el diff contra `boe.py` sigue garantizando la paridad de la herramienta.
- El nombre `boe-legislacion` es el que ya usan los documentos semilla y el catálogo de skills. Es un
  directorio: si al especificar H5 se prefiere otro, se cambia allí sin coste.

No se reabre ninguna decisión cerrada: «skills primero» (ADR 0008), «skills sin código», genericidad
territorial (ADR 0009), el multicall y el sobre de salida siguen igual. Cambia qué skill cierra la fase 0
y cuándo llegan las verticales.

## Consecuencias

**A favor**

- La primera skill sirve a cualquier consulta de normativa, y las skills municipales de las fases 1 a 3
  tienen desde H5 una base que citar.
- El orden del roadmap coincide con el del producto: base genérica primero, especialización después y
  poco a poco, medible con evals en cada paso.
- H5 se mantiene pequeño y sin renumerar nada; la migración de `boe-fiscal` deja de bloquear y de estar
  bloqueada, y se hace cuando aporte valor.
- YAGNI gana un criterio más: ni campo `vertical` ni packs por vertical hasta la primera vertical.

**En contra, y asumido**

- `boe-fiscal` sigue en Python fuera del repo hasta su hito de vertical. No se pierde nada: hoy funciona
  y se sigue usando igual.
- El ADR 0008 afirmaba que «la fase 0 no cambia»; queda anotado allí como superado en ese punto.
- `refs/mapa-sistema-legal-skills.md` y `refs/kitlegal-estructura-y-ecosistema.md` siguen presentando
  `boe-fiscal` como primera skill y `boe-legislacion` como su generalización posterior. Son documentos
  semilla y no se reescriben; `refs/00-README.md` avisa de la lectura correcta y, ante conflicto,
  prevalecen la constitución y el roadmap.
