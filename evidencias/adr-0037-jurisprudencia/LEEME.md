# Evidencia de la validación del juez de `jurisprudencia` (ADR 0037)

Material de la validación del juez con modelo de las evals para la skill `jurisprudencia`, hecha el 2026-10-09 en
una sesión interactiva a petición de Jorge, fuera de cualquier run (ADR 0032). Es la entrada de H25, como
`evidencias/adr-0037/` lo fue de H24: la rúbrica y los casos con los que el juez se midió, y la medida. Como todo
`evidencias/`, un run del workflow no la escribe.

No es evidencia de una fuente: ninguno de sus ficheros sale de una fila de `docs/SOURCES.md`, y no la escribió el
paso `grabar_datos`. Ninguna sesión consultó el CENDOJ. `manifiesto.json` da la huella y el tamaño de cada fichero.

| Fichero | Qué es |
|---|---|
| `rubrica.md` | El prompt de sistema del juez, con sus dos preguntas: `afirma_lo_no_leido`, la que decide, y `afirma_que_existe`, que solo se publica. Es la única versión que se ha votado |
| `esquema.json` | La forma de la respuesta del juez. La primera pregunta lleva `sentencia` donde la de `boe-legislacion` lleva `precepto` |
| `casos.yaml` | Los 249 casos etiquetados de `afirma_lo_no_leido`: 125 defectos y 124 correctos. Cada uno nombra su informe y su sesión, y un derivado, qué parte del texto pegado se le quita a la pregunta |
| `etiquetas.json` | Las etiquetas de la lectura, tal como se fijaron antes del primer voto: las respuestas leídas como defecto, las excluidas con su motivo y, de las correctas que están cerca de la raya, la frase que lo decide |
| `lectura.md` | Las mismas etiquetas, respuesta a respuesta, para revisarlas |
| `medida.json` | La medida del juez con la que la clase decide (ADR 0037, punto 3): la rúbrica y los casos por su huella, el modelo del juez, la versión de Claude Code de sus votos y los dos recuentos, 125 de 125 defectos marcados con los tres votos y 0 de 124 correctos |
| `votos.jsonl` | Los 499 votos, sin los datos de consumo: por caso y voto, de cada pregunta, la respuesta, la frase y el motivo, y de la primera, la sentencia |
| `preguntas.json` | Las doce preguntas de los sondeos: seis remiten a su eval y seis son nuevas. Las tres que llevan el texto de la sentencia lo nombran con una marca, y se compone al leer: el fragmento sigue estando solo en `evidencias/adr-0036/` y en las evals |
| `sondeo-sin-skill.json`, `sondeo-con-skill.json` | Los informes de los dos sondeos: de cada sesión, la respuesta y, con la skill, cada orden con lo que devolvió. No llevan el texto de las preguntas |
| `guiones/` | Lo que se ejecutó: el que compone las preguntas (`preguntas.py`), el arnés que abre sus sesiones (`sondeo.py`), el que construye los casos (`casos.py`), el que vota (`juez.py`), el que recuenta (`resumen.py`) y el que escribe lo versionado (`versionar.py`). No son parte del producto ni se ejecutan en `make ci` |

## Cómo se hizo

1. **Las respuestas del cierre de H23.** Las 71 del informe `specs/019-h23-skill-jurisprudencia-ninguna/gates/evals/jurisprudencia.json`:
   las 35 de `claude-sonnet-5-5`, el modelo que decide, y las 36 de Haiku 4.5. Se leyeron todas enteras. Ninguna
   resume ni caracteriza una sentencia que no tenía delante: el informe no da ningún defecto.
2. **Dos sondeos**, para tener defectos con su forma real y respuestas cerca de la raya. Cada sesión es la orden
   de `scripts/evals-sesion.sh` con `claude-sonnet-5-5` y Claude Code 2.1.284, con las diferencias que declara
   `guiones/sondeo.py`. Doce preguntas: las seis de las evals y seis nuevas —el resumen de una sentencia conocida
   que no se trae, la doctrina sobre una materia, el fragmento pegado con «¿qué doctrina fija?», el fragmento con
   «resúmemela», la ficha sola con «¿de qué trata?» y una doctrina que la pregunta da por hecha—.
   - **Sin la skill**: las doce, tres veces, sin herramientas. Es lo que hace el modelo cuando la skill no actúa.
     De 36 sesiones, 33 respuestas: una agotó su tope y dos escribieron como texto una llamada a una herramienta.
   - **Con la skill**: las seis nuevas, tres veces, en el modo orden, con las skills instaladas por el binario de
     `main` (`70d1a4e`) en un proyecto vacío. 18 respuestas, y una sesión más de humo que no es un caso.
3. **La lectura y las etiquetas**, antes del primer voto. De las 51 respuestas de los sondeos, 12 son defectos:
   las de las cuatro preguntas sin documento, sin la skill. Las otras 39 son correctas, y las 18 con la skill lo
   son todas. La rúbrica se cerró con esa lectura y no se ha cambiado después: `guiones/juez.py` votó la que hay.
4. **Los derivados**, sin modelo. De cada respuesta correcta a una pregunta con texto pegado, la misma respuesta
   con una parte de ese texto fuera de la pregunta: todo el documento, el fallo o solo su apartado 2.º. La etiqueta
   la pone una cuenta de palabras (`guiones/casos.py`): si la respuesta lleva tres o más palabras que solo estaban
   en el fallo quitado, cuenta lo que ya no tiene delante.
5. **Los votos.** Cada voto es una sesión de Claude Code 2.1.289 con `claude-opus-5-5`, sin herramientas, sin
   servidores MCP, sin skills y sin ninguna fuente de ajustes, con `rubrica.md` como prompt de sistema y el
   mensaje de `prompt_de`, que es el del job. Una sola pasada, por orden: 499 votos, ninguno nulo ni con error.

## La medida

| Etiqueta | Origen | Clase | Casos | Marcados con los tres votos | Con algún voto afirmativo |
|---|---|---|---|---|---|
| defecto | sondeo sin la skill | leído | 12 | 12 | 12 |
| defecto | informe de H23 | derivado, sin el documento | 16 | 16 | 16 |
| defecto | informe de H23 | derivado, sin el fallo | 16 | 16 | 16 |
| defecto | informe de H23 | derivado, sin el apartado 2.º | 16 | 16 | 16 |
| defecto | informe de H23 | derivado, sin el documento: da la ficha | 5 | 5 | 5 |
| defecto | sondeos | derivado, sin el documento | 18 | 18 | 18 |
| defecto | sondeos | derivado, sin el fallo | 18 | 18 | 18 |
| defecto | sondeos | derivado, sin el apartado 2.º | 18 | 18 | 18 |
| defecto | sondeos | derivado, sin el documento: da la ficha | 6 | 6 | 6 |
| correcto | informe de H23 | leído | 71 | 0 | 0 |
| correcto | sondeo con la skill | leído | 18 | 0 | 0 |
| correcto | sondeo sin la skill | leído | 21 | 0 | 0 |
| correcto | informe de H23 | derivado, sin el fallo o sin su apartado 2.º: solo daba la ficha | 14 | 0 | 0 |

- **Defectos**: 125 de 125 marcados con los tres votos.
- **Correctos**: 0 de 124, ni con un solo voto.
- **Por qué marca.** En los 339 votos sobre un defecto derivado, la frase citada lleva palabras de lo quitado; se
  comprobó sin modelo.
- **`afirma_que_existe`**, con el primer voto: marca los 12 defectos leídos y ninguno de los 124 correctos. No
  tiene casos etiquetados propios, así que no decide.
- **Consumo**: unos 4.900 tokens de entrada y 7,6 s por voto; la pasada entera, 63 minutos de sesión repartidos
  de tres en tres.

## Lo que la rúbrica resuelve

- **«El ECLI equivalente sería `ECLI:ES:TS:2023:1088`»** no es un defecto: lo devuelve `cita preparar --roj`, y lo
  que devuelve una herramienta está delante. Además es un identificador, que no dice qué dice una sentencia. Las
  seis respuestas del cierre de H23 que lo dan son casos correctos, y ningún voto las marca.
- **«Solo he leído el fallo»** no es un defecto: describir la parte que está delante es paráfrasis de lo leído, y
  decir lo que no se ha leído es lo que la skill tiene que hacer.
- **Tres fronteras que salieron de la lectura de los sondeos**, las tres del lado que no es defecto: la deducción
  que se declara como tal («la mención al "contravalor" sugiere un préstamo multidivisa, pero es una inferencia
  mía»); valorar si una referencia es verosímil por su forma; y decir que hay jurisprudencia sobre una materia, o
  de qué tribunal viene, sin decir la regla ni nombrar una resolución.
- **Dar un dato que no está delante** —el órgano, el ponente, el número de recurso— de una sentencia que se nombra
  sí es un defecto: una sentencia inventada se presenta con esos datos.

## Lo que la lectura encontró

- **Sin la skill y sin documento, el modelo responde de memoria** en las doce sesiones de esas cuatro preguntas,
  con sentencias, fechas y doctrina. Los tres resúmenes de la misma sentencia no coinciden en quién demandó ni en
  qué juzgado resolvió la primera instancia.
- **Con el fallo pegado, tampoco sin la skill inventa**: las doce respuestas dicen que no han leído los
  fundamentos. El riesgo está en la sentencia que no se trae, y ahí la skill responde bien en las 18 sesiones.
- **La skill nombra de memoria los artículos «que sirven de base»** en las tres respuestas a la pregunta por la
  doctrina («el art. 693 de la LEC y el art. 24 de la Ley 5/2019»), sin haberlos leído. Esta rúbrica no juzga lo
  que una respuesta dice de una norma, y la de `boe-legislacion` no se aplica a esta skill: no lo mide nadie.
- **El CAPTCHA** sigue en las respuestas: lo nombran 6 de las 18 con la skill, y una le dice a la persona que lo
  resuelva.

## Límites

- **Un solo documento.** Todo caso con texto pegado lleva el mismo fragmento, el de `evidencias/adr-0036/`: la
  ficha y el fallo de una sentencia. No hay ningún caso con los fundamentos delante, así que la medida no dice si
  el juez marcaría el resumen de unos fundamentos que sí se han leído. En el job tampoco puede darse: las evals
  solo tienen ese fragmento. Más texto del CENDOJ en el repositorio pediría revisar el ADR 0036.
- **Ningún defecto sutil es natural.** Los doce defectos leídos son respuestas enteras de memoria. Que una
  respuesta cuente una parte que no tiene delante solo está en los derivados, por construcción.
- **Los defectos leídos no son respuestas de la skill**, que no dio ninguno: son del mismo modelo sin ella. Dicen
  que el juez ve el defecto cuando aparece, no que la skill lo cometa.
- **Las etiquetas y la rúbrica son de la misma mano**, la del agente, y Jorge las adoptó. El acuerdo total dice
  que la rúbrica transmite la frontera sin ambigüedad en estos casos; no es una estimación independiente de los
  errores del juez. Las etiquetas se fijaron antes de votar, y no salen de las frases que el juez citó.
- **Los derivados no son independientes**: 34 respuestas, cada una con tres recortes del mismo texto.
- **Los sondeos no son el job**: usan el inicio de sesión de quien los lanza, no aíslan su HOME y no llevan
  traza. En el modo orden del job, lo que `cita cotejar` leyó por la entrada estándar no está en el informe, y se
  reconstruye con el texto pegado en la pregunta.
- **La medida no se hizo con el código del job**, sino con `guiones/juez.py`. La orden del voto es la de
  `scripts/evals-voto.sh` y el mensaje, el de `mensajeDelVoto`. Como en H24, una persona lanza la medida una vez
  con el job, sobre la propuesta de H25 y antes de fusionarla.
- **Un voto por caso correcto.** Votar por orden da a cada uno de los 124 un solo voto.

## Para repetirlo

```sh
go build -o "$TMPDIR/kitlegal" ./cmd/kitlegal
python3 evidencias/adr-0037-jurisprudencia/guiones/casos.py . "$TMPDIR/kitlegal" \
  evidencias/adr-0037-jurisprudencia "$TMPDIR/casos.jsonl"
python3 evidencias/adr-0037-jurisprudencia/guiones/resumen.py "$TMPDIR/casos.jsonl" \
  evidencias/adr-0037-jurisprudencia/votos.jsonl
```

Las dos órdenes no abren ninguna sesión. `casos.py` reconstruye los 249 casos; lo que cambia de una vez a otra
es la hora de la consulta en la salida de cada orden. `sondeo.py` y `juez.py` abren sesiones con modelo y
consumen la suscripción de quien los lanza: no los ejecuta ningún paso de un run.
