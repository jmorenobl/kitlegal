# Evidencia del ADR 0037

Material de la validación del juez con modelo de las evals (ADR 0037, «Validación»), hecha el 2026-10-04 y el
2026-10-05 en una sesión interactiva a petición de Jorge, fuera de cualquier run (ADR 0032). Es la entrada de H24:
la rúbrica y los casos con los que el juez se midió. Como todo `evidencias/`, un run del workflow no la escribe.

No es evidencia de una fuente: ninguno de sus ficheros sale de una fila de `docs/SOURCES.md`, y no la escribió el
paso `grabar_datos`. `manifiesto.json` da la huella y el tamaño de cada fichero.

| Fichero | Qué es |
|---|---|
| `rubrica.md` | El prompt de sistema del juez, con sus dos preguntas: `afirma_lo_no_leido` y `cuenta_su_proceso`. Es la cuarta versión, la de la frontera del precepto identificado |
| `esquema.json` | La forma de la respuesta del juez |
| `casos.yaml` | Los 259 casos etiquetados de `afirma_lo_no_leido`: 5 defectos de la bitácora, 140 derivados y 114 de la lectura (67 defectos y 47 avisos correctos). Cada uno nombra su informe versionado y su sesión |
| `lectura.md` | Las 114 etiquetas de la lectura, con la frase que decide cada una, para revisarlas |
| `votos-ajuste.jsonl`, `votos-medida.jsonl` | Los 1.189 votos de la medida final, sin los datos de consumo: por caso y voto, la respuesta, la frase, el precepto y el motivo |
| `guiones/` | Lo que se ejecutó: el test que reconstruye los textos (`arnes_test.go.txt`), y los guiones que construyen los casos, votan y resumen. No son parte del producto ni se ejecutan en `make ci` |

## Cómo se hizo

1. **Los textos.** De cada respuesta del modelo que decide, en las evals que activan la skill de los seis informes
   de `specs/01{1..6}-*/gates/evals/boe-legislacion.json` (423 respuestas), se reconstruyeron sin modelo los textos
   que devolvieron sus herramientas: se prepara la caché y el grafo previo con el código del job y se repite cada
   orden de la sesión con `--offline`.
2. **Los casos.** Las 423 respuestas, más 140 derivadas: por informe, eval y modo, la primera sesión que pasó, sin
   el texto del primer bloque que cita.
3. **Los votos.** Cada voto es una sesión de Claude Code 2.1.289 con `claude-opus-5-5`, sin herramientas, sin
   servidores MCP, sin skills y sin ninguna fuente de ajustes, con `rubrica.md` como prompt de sistema.
4. **Cuatro vueltas**, con cuatro rúbricas. Solo la última está aquí; las cifras de todas están en el ADR.

## Límites

- **Las etiquetas de la lectura y la rúbrica son de la misma mano.** El acuerdo total de la medida final dice que
  la rúbrica transmite la frontera sin ambigüedad; no es una estimación independiente de los errores del juez.
- **Las etiquetas de la lectura salen de las frases que el juez citó** en alguna vuelta. Una respuesta etiquetada
  como correcta puede llevar en otra parte una frase que ningún voto citó.
- **Los textos reconstruidos son un superconjunto** de lo que la sesión vio, si filtró la salida de una orden. En el
  job, el juez recibe los textos del transcript.
- **Las 303 respuestas que ninguna rúbrica marcó** no se han leído enteras.
