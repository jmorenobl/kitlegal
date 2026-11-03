---
name: articulos-del-boe
description: >-
  Lee y cita con kitlegal un artículo de una norma consolidada del Boletín Oficial del Estado (BOE), de cualquier
  materia: busca su bloque en el índice de la norma y lee su texto vigente. Se usa siempre que la pregunta pida qué
  dice un artículo de una ley, de un real decreto o de la Constitución, también cuando el agente crea saberlo. Se
  activa antes de boe_indice o boe_articulo, o de kitlegal boe indice o kitlegal boe articulo: qué pedir y cómo
  responder lo dice la skill.
---

# Artículos del BOE: leer y citar un artículo

Esta skill lee con kitlegal el texto vigente de un artículo de una norma consolidada del BOE y responde con él; nada de
lo que diga la respuesta sale de memoria: sale del texto que devuelve la herramienta.

## Cómo se lee un artículo

1. La norma, por su identificador del BOE, `BOE-A-<año>-<número>`. Los de cuatro normas: Constitución Española,
   BOE-A-1978-31229; Ley 58/2003, General Tributaria, BOE-A-2003-23186; Ley 40/2015, de Régimen Jurídico del Sector
   Público, BOE-A-2015-10566; Estatuto de los Trabajadores, BOE-A-2015-11430.
2. El bloque del artículo, en el índice de la norma: `kitlegal boe indice <norma>` o la herramienta `boe_indice`. El
   bloque es el campo `id` de la entrada del artículo.
3. El texto del artículo: `kitlegal boe articulo <norma> <bloque>` o la herramienta `boe_articulo`.

## Cómo se responde

Lo que dice el artículo, con las palabras del texto leído y sin nada que no esté en él; cita la norma y el bloque así:
[BOE-A-1978-31229, bloque a140].

Antes de escribir la respuesta, lee [references/cierre.md](references/cierre.md): dice cómo termina.
