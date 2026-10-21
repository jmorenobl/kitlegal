---
id: H36
fase: "1"
---

# H36 · El motor de cómputo descrito por datos

- **Objetivo**: que un plazo de otra materia entre como datos con sus preceptos, no como una rama nueva del código. Hoy hay dos cómputos fijos (`lpac` y `contencioso`), todo lo que no es contencioso se cuenta por la LPAC y el sábado y el domingo son siempre inhábiles.
- **Entrega**: las reglas de cómputo en `data/plazos.yaml`, cada una con su precepto: unidad, días que excluye, periodos inhábiles, día inicial, prórroga y meses de fecha a fecha; `lpac` y `contencioso` como datos, con los goldens de H9 idénticos; la regla general del art. 5 del Código Civil como segunda implementación, con el precepto leído con `kitlegal boe articulo` antes de fijarla; los tipos de plazo, fuera del código. H9.1 añade después sus cómputos como datos.
- **Riesgo**: parametrizar de más: solo los ejes que piden los preceptos de los cómputos que existen y el del Código Civil.
- **Controles**: los goldens y los property tests de H9; un cómputo de ensayo que entra sin tocar Go.
