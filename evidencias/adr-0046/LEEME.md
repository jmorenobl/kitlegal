# Evidencia de la validación del juez de las descripciones (ADR 0046)

Material de la validación de la clase `habla_en_primera_o_segunda_persona`, la que juzga si la descripción de una
skill está en tercera persona. Se hizo el 2026-10-23 en una sesión interactiva, a petición de Jorge y fuera de
cualquier run (ADR 0032). Es la entrada de H31, como `evidencias/adr-0037-jurisprudencia/` lo fue de H25. Como todo
`evidencias/`, un run del workflow no la escribe.

No es evidencia de una fuente: ningún fichero sale de una fila de `docs/SOURCES.md`, y ninguna sesión consultó la
red, salvo la del modelo.

| Fichero | Qué es |
|---|---|
| `rubrica.md` | El prompt de sistema del juez, con su única pregunta. Es la única versión que se ha votado |
| `esquema.json` | La forma de la respuesta del juez |
| `historicas.json` | Las ocho descripciones distintas que han tenido las skills del repositorio en su historial, con la skill y el commit: las de `boe-legislacion`, `legal-core` y `jurisprudencia`, y la de `seo-semanal` |
| `casos.yaml` | Los 99 casos etiquetados: 63 defectos y 36 correctos, con la forma que hace defecto a cada uno |
| `medida.json` | La medida con la que la clase decide: la rúbrica y los casos por su huella, el modelo del juez, la versión de Claude Code de sus votos y los dos recuentos |
| `votos.jsonl` | Los 225 votos: por caso y voto, la respuesta, la frase, el motivo, la duración y el modelo |
| `guiones/` | Lo que se ejecutó: el que construye los casos (`casos.py`), el que vota (`juez.py`) y el que recuenta (`resumen.py`). No son parte del producto ni se ejecutan en `make ci` |

## Cómo se hizo

1. **Las descripciones leídas.** Las ocho del historial dicen «Úsala», y tres de ellas, además, «Actívala» y «creas»:
   son los ocho defectos leídos.
2. **Las bases.** De cada una, una base en tercera persona con cinco sustituciones fijas: «Úsala siempre que» y
   «Úsala cuando» por «Se usa…», «creas conocer…» por «el agente crea conocer…» y «Actívala antes de llamar a» por
   «Se activa antes de llamar a». `casos.py` comprueba que no queda ninguna de las tres formas.
3. **Los derivados**, sin modelo. De cada base, los que meten una forma de primera o de segunda persona son defectos:
   el imperativo de tú y de usted, con pronombre pegado («Úsala», «Actívala», «Úsela»), el subjuntivo («creas»), el
   pronombre y el posesivo de tú («te preguntan», «tu pregunta», «tu respuesta»), el de vosotros («os piden») y la
   primera del singular y del plural («Te ayudo», «nuestras»). Los que la meten dentro de una cita, o la cambian por
   una forma impersonal o de tercera persona, son correctos: dos preguntas citadas («¿hasta cuándo puedo recurrir?»,
   «¿puedes decirme…?»), «Conviene activarla», «El agente la activa antes…» y un verbo que se escribe como un
   imperativo de tú con la skill de sujeto («Lee la norma con kitlegal y responde citando…»). La etiqueta la pone la
   construcción, antes del primer voto; un derivado igual a otro caso no se repite.
4. **Los votos.** Cada voto es una sesión de Claude Code 2.1.289 con `claude-opus-5-5`, sin herramientas, sin
   servidores MCP, sin skills y sin ninguna fuente de ajustes, con `rubrica.md` como prompt de sistema y el mensaje
   de `mensaje_de`: la descripción entre `<descripcion skill="…">` y `</descripcion>`, y la petición. Una sola
   pasada, por orden y de tres en tres: 225 votos, ninguno nulo ni con error, 3,9 s de mediana y 9 s como mucho.

## La medida

| Etiqueta | Clase | Casos | Marcados con los tres votos | Con algún voto afirmativo |
|---|---|---|---|---|
| defecto | leída | 8 | 8 | 8 |
| defecto | imperativo de tú | 4 | 4 | 4 |
| defecto | imperativo de tú: activar | 3 | 3 | 3 |
| defecto | imperativo de usted | 8 | 8 | 8 |
| defecto | subjuntivo de tú | 4 | 4 | 4 |
| defecto | pronombre de tú | 8 | 8 | 8 |
| defecto | posesivo de tú | 4 | 4 | 4 |
| defecto | pronombre de vosotros | 8 | 8 | 8 |
| defecto | primera del singular | 8 | 8 | 8 |
| defecto | primera del plural | 8 | 8 | 8 |
| correcto | base en tercera persona | 8 | 0 | 0 |
| correcto | primera y segunda dentro de una cita | 8 | 0 | 0 |
| correcto | impersonal | 8 | 0 | 0 |
| correcto | el agente en tercera persona | 8 | 0 | 0 |
| correcto | verbo que se escribe como un imperativo, con la skill de sujeto | 4 | 0 | 0 |

- **Defectos**: 63 de 63 marcados con los tres votos.
- **Correctos**: 0 de 36, ni con un solo voto.
- **Por qué marca.** En los 63 defectos, alguna de las frases citadas contiene la forma que hace defecto al caso; se
  comprobó sin modelo.

## Límites

- **Ningún defecto sutil es natural.** Los ocho leídos dicen «Úsala» al principio de una frase. Los demás son
  derivados, por construcción, y la mitad añaden una frase entera al final.
- **Los derivados no son independientes**: salen de ocho descripciones, cuatro de ellas de `boe-legislacion`.
- **Solo español.** Ningún caso está en inglés, y la rúbrica no se ha medido con «Use when…».
- **Las etiquetas y la rúbrica son de la misma mano**, la del agente. El acuerdo total dice que la rúbrica transmite
  la frontera en estos casos; no es una estimación independiente de los errores del juez.
- **La medida no se hizo con el código del job**, sino con `guiones/juez.py`, con la orden del voto de
  `scripts/evals-voto.sh`. Como en H24 y H25, una persona la lanza una vez con el job, con la etiqueta
  `evals-medir-juez`, sobre la propuesta de H31 y antes de fusionarla.

## Para repetirlo

```sh
python3 -I evidencias/adr-0046/guiones/casos.py evidencias/adr-0046/historicas.json "$TMPDIR/casos.jsonl"
python3 -I evidencias/adr-0046/guiones/resumen.py "$TMPDIR/casos.jsonl" evidencias/adr-0046/votos.jsonl
```

Las dos órdenes no abren ninguna sesión. `juez.py` abre sesiones con modelo y consume la suscripción de quien lo
lanza: no lo ejecuta ningún paso de un run.
