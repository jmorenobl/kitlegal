---
id: H44
fase: "1"
depende_de: [H35, H42]
---

# H44 · La doctrina administrativa como capacidad, con la DGT y el TEAC (ADR 0059)

- **Objetivo**: que lo que dice una norma tributaria se complete con cómo la interpreta la Administración, con el
  documento delante. Una consulta de la DGT o una resolución del TEAC se cita leída, nunca de memoria, y se distingue de
  la norma. Una consulta vinculante obliga a los órganos de la Administración tributaria en los términos del art. 89
  LGT, que se lee antes de fijarlo; lo que no es vinculante se dice. Es la capacidad de doctrina de la base, validada con
  dos implementaciones reales.
- **Antes de escribir la entrada**: las condiciones de PETETE (DGT) y DYCTEA (TEAC), revisadas por una persona en
  `docs/SOURCES.md` (ADR 0059, punto 3). Lo que digan decide la vía:
  - si un programa puede leer un documento por su identificador, un applet lo trae con el sobre de siempre, y si puede
    buscar, busca;
  - si no, el patrón de `jurisprudencia` (ADR 0036): la skill prepara la búsqueda exacta, la persona busca con su
    navegador y trae el documento, y el binario resuelve el identificador y lo coteja con lo de H35.
- **Entrega**: la capacidad de doctrina —una skill o una referencia; lo decide el spec con el ADR 0048— con las reglas
  de `jurisprudencia`: nada se cita ni se caracteriza sin leerlo, y la doctrina se distingue de la norma y de la
  sentencia. Las dos fuentes, la DGT y el TEAC, entran como datos de la capacidad. La referencia tributaria de H42 dice
  cuándo pedir doctrina.
- **Alcance**: fuera, la doctrina de otras materias, que llega con su paquete sobre esta capacidad, y la que publica el
  BOE (H28). Nada de PETETE ni de DYCTEA se lee sin sus condiciones revisadas: sin búsqueda automática ni descarga
  masiva si no las admiten.
- **Controles**: los de una fuente nueva si hay applet (fixtures, golden, `verify-sources`), o golden del cotejo de un
  documento que es el pedido y de otro que no lo es; evals de una pregunta tributaria con su consulta vinculante, juzgadas
  por un juez medido contra respuestas reales antes de decidir (ADR 0037).
- **Aceptación**: «¿cómo interpreta Hacienda X?» se responde con la consulta o la resolución leída y citada, diciendo si
  vincula, o con la búsqueda preparada y el cotejo del documento que trae la persona.
