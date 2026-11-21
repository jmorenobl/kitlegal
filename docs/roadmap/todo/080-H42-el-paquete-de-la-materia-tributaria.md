---
id: H42
fase: "1"
depende_de: [H36]
---

# H42 · El paquete de la materia tributaria, la segunda materia real de la base (ADR 0059)

- **Objetivo**: que el molde del paquete de materia (H32) se pruebe con una segunda materia real, no con la vertical de
  ensayo de `testdata/`, y que un paquete saque por primera vez lo suyo de lo que la ley regula aparte. Una pregunta
  tributaria de lo que el paquete cubre se responde con la norma citada y el plazo del binario. La que queda fuera se
  declara con la línea de siempre.
- **Entrega**:
  - las normas tributarias con `vertical: fiscal` en `data/normas.yaml`: la LGT, la del IRPF y el TRLRHL, que ya tienen
    evals de `normativa`, y las del IVA y el IS si el spec lo justifica;
  - una referencia de la materia en `plazos-y-territorio` y otra en `normativa`;
  - sus filas de `cubierto` con `materia: fiscal`;
  - los plazos de la revisión en vía administrativa de un acto tributario —el recurso de reposición (art. 223 LGT) y la
    reclamación económico-administrativa (art. 235 LGT), con cada precepto leído con `kitlegal boe articulo` antes de
    fijarlo— como datos del motor de H36;
  - lo que salga de `regulados_aparte` por quedar cubierto, y nada más.
  La doctrina de la DGT y del TEAC llega con H44, sobre este paquete.
  Todo entra sin tocar Go ni el protocolo de ninguna skill (principio X).
- **Alcance**: es el paquete de materia, no la vertical fiscal (ADR 0059, punto 2). Quedan fuera, y siguen declarados
  no cubiertos cuando ya lo están:
  - el calendario del contribuyente y los plazos de fecha fija;
  - los modelos y las autoliquidaciones;
  - el sancionador tributario (DA 1ª.2.c);
  - las aduanas;
  - la aplicación de los tributos fuera de la revisión;
  - las haciendas forales;
  - los escritos;
  - los órganos y destinatarios, salvo que H37 tome de aquí su segunda implementación.
- **Riesgos**:
  - **Los tests con listas literales.** Hay tests que fijan con listas cerradas lo que lleva `data/`, como el
    comentario de `data/normas.yaml` sobre «las diez de lo administrativo». Si el hito los cambia a mano en lugar de
    hacerlos derivar de `data/`, «sin tocar código» deja de ser verdad.
  - **La norma de lo que sigue aparte.** Al sacar una fila, lo que sigue regulado aparte no expone su norma (ADR 0058,
    punto 2).
  - **El coste de las evals.** Etiquetar normas dispara las evals de las skills que las leen, entre ellas la 04 de
    `normativa`; como mucho dos sondeos por pregunta.
- **Controles**:
  - evals en `evals/` de las skills que cambian: un plazo cubierto de la revisión con su fecha del binario, un
    procedimiento tributario que sigue no cubierto con su línea, una pregunta sustantiva con la norma citada y una de
    un municipio de régimen foral declarada no cubierta;
  - las evals de lo administrativo, sin cambios y en verde;
  - un test de que el paquete entra sin cambios en Go ni en ningún `SKILL.md`.
- **Aceptación**: «me notificaron una liquidación del IRPF el día D: ¿hasta cuándo puedo recurrirla en reposición?» da
  la fecha del binario con su precepto; «¿cuánto tengo para recurrir una sanción tributaria?» se declara no cubierto,
  sin plazo ni vía.
