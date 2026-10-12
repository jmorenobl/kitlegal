# 0038 · El umbral de una clase de significado es una tasa, fijada desde lo medido; el cero se queda en lo que tiene forma

- **Estado**: aceptada (2026-10-12). El job sigue con el umbral anterior hasta que lo cambie el hito que lo
  implemente («Enmiendas que provoca»).
- **Fecha**: 2026-10-12
- **Hito**: transversal, tras el run de H25 y antes de fusionar su propuesta (#135). Sustituye, del ADR 0037, el
  umbral del punto 7 de su decisión —`afirma_lo_no_leido` «decide, con umbral 0»— y la frase de su validación que
  daba por hecho que la skill llegaría a cero. El resto del ADR 0037 sigue vigente: el juez, la regla de los votos,
  la frontera del precepto identificado, la medida del juez y dónde corre. El ADR 0029 sigue vigente entero: el
  umbral que aquí se decide hace fallar el job.

## Contexto y problema

`afirma_lo_no_leido` de `boe-legislacion` decide con umbral 0 sobre las 54 respuestas de cada modo. El ADR 0037 lo
fijó así contando con que H24 corregiría la skill («con el umbral en 0 saldría en rojo por la skill, no por el
juez: H24 la corrige»). H24 la mejoró y no la llevó a cero:

| Medición | Marcadas |
|---|---|
| Informes anteriores a H24, con la rúbrica final (ADR 0037): H7.4, H21 y H22 | 3 de 54, 11 de 108 y 6 de 108 |
| Cinco mediciones de cierre con v0.1.7 y v0.1.8: la última de H24, la de H23 y las tres de H25 | 5 de 536 (0,9 %; entre 0,3 % y 2,2 %) |

Con una tasa del 0,9 %, una ejecución de 108 respuestas sale limpia el 36 % de las veces. El cierre de un hito mide
hasta tres veces, así que pasa el 74 % de los cierres sin que la tasa haya cambiado: H24 pasó a la tercera. Y el
cierre de H25, que no tocaba `boe-legislacion`, salió en rojo en sus tres mediciones por ella.

Las marcas son buenas según la rúbrica, y casi todas están en las preguntas que no nombran el artículo (evals 15 y
16). Sobre esas dos evals se ha probado, fuera de un run y con el juez del job, el 2026-10-11 y el 2026-10-12
(`evidencias/adr-0038/`; `docs/USO.md`, 2026-10-12):

| Condición | Marcadas |
|---|---|
| `claude-sonnet-5-5`, skill de `main` (eval 15) | 4 de 30 |
| `claude-sonnet-5-5`, con la rúbrica de cada artículo en el índice (eval 15) | 6 de 30 |
| `claude-opus-5-5`, skill de `main` (evals 15 y 16) | 3 de 60 |
| `claude-sonnet-5-5`, con un paso de repaso antes de responder (evals 15 y 16) | 2 de 60 |

Además, la reparación del cierre de H25 añadió a la skill, entre lo prohibido, la forma «el art. N sobre
<materia>», y la medición siguiente trae «el art. 27 sobre tipicidad».

Ni más material, ni un repaso, ni más prohibiciones, ni el modelo más capaz acercan la tasa a cero. Lo que queda
son frases como «las exenciones del art. 7» o «no puedo precisar qué parte de esas rentas queda exenta». El texto de
una respuesta lo genera un modelo, y esa generación no se controla: se puede reducir lo que falla, no garantizar
que no falle nunca.

Un umbral de cero sobre algo así no es un control. Su verde no dice que la skill cumpla, porque un 0,9 % lo pasa
una de cada tres veces, y su rojo no dice nada del cambio que se mide, porque depende del sorteo. Empuja además a
las dos cosas que el proyecto no quiere: volver a medir hasta que salga limpio, y reparar a ciegas una skill que el
hito no tocaba.

## Opciones consideradas

1. **Seguir con el cero y buscar el arreglo.** Rechazada. Cuatro condiciones medidas no lo dan, y mientras tanto
   cada cierre es un sorteo.
2. **Que decida un modelo más capaz y Sonnet solo se publique.** Rechazada. Opus da 3 de 60 en las mismas
   preguntas: con cero, su cierre saldría en rojo más o menos la mitad de las veces. Además, la mayoría del uso real
   no va con Opus (ADR 0016).
3. **Que el cierre decida solo con las skills que el hito cambia.** Rechazada. Deja el umbral solo publicado para
   casi todos los hitos, que es lo que el ADR 0029 corrigió, y no arregla el cierre del hito que sí toca la skill.
4. **Sacar del umbral las evals informativas**, donde están casi todas las marcas. Rechazada. Preguntar por materia
   es como pregunta la gente, y el ADR 0029 ya decidió que un umbral agregado las cuenta.
5. **Mover la frontera de la rúbrica** para que una etiqueta breve no cuente. No se decide aquí. La validación del
   ADR 0037 encontró tres etiquetas de memoria que estaban mal, y cambiar la rúbrica pide reetiquetar los casos y
   repetir la medida del juez. Queda para cuando haya más cierres medidos.
6. **Publicar la tasa sin que decida.** Rechazada: publicar no es un control (ADR 0029).
7. **Un umbral en forma de tasa, fijado desde lo medido, que hace fallar el job.** Elegida.

## Decisión

1. **Cero donde hay forma, tasa donde hay significado.** Lo que un guion comprueba —la cita con su identificador
   y su bloque, las líneas `⚠`, la activación, `cita_sin_documento`, la red y la medida del juez— sigue con umbral
   0: es un hecho de la sesión, y se cumple. Una clase del juez que decide lleva un umbral en forma de tasa.
2. **El umbral de `afirma_lo_no_leido` en `boe-legislacion`**: como mucho 3 respuestas marcadas de las 108 del
   modelo que decide, con los dos modos juntos (`umbral: 0.028` en `clases.yaml`). El recuento de cada modo se
   sigue publicando, sin decidir.
3. **La regla del número.** El umbral de una clase es el más estricto con el que una ejecución, a la tasa
   acumulada de la skill, sale en rojo sin motivo como mucho 1 de cada 50 veces. Con 108 respuestas: 3 si la tasa
   es del 0,95 % o menos, 2 si es del 0,53 % o menos y 1 si es del 0,20 % o menos. La tasa acumulada es la de los
   cierres versionados con la skill como está, sobre 500 respuestas al menos.
4. **El umbral solo se estrecha.** Cuando la tasa acumulada baja de un escalón, una persona lo baja con un diff de
   `clases.yaml` y sus cifras. Ningún run lo cambia y ningún corrector lo sube (ADR 0029); subirlo pediría otro ADR.
5. **`jurisprudencia` nace con la misma forma**: como mucho 1 marcada de las 60 respuestas del modelo que decide.
   Ha medido 0 de 180, que no distingue una tasa de cero de una del 1,6 %; con cero como umbral, una tasa real del
   0,5 % pondría en rojo 1 de cada 4 ejecuciones. Su cero sin modelo, `cita_sin_documento`, no cambia.
6. **La serie se versiona.** Cada cierre añade una fila a la serie de la skill: marcadas y total, por modo y por
   clase de pregunta —la que nombra el artículo y la que pregunta por materia—, con la versión de la skill y el
   modelo. El informe final de un hito da la serie y la tasa acumulada. Es la medida con la que se aplican los
   puntos 3 y 4, y la que dice si un cambio de la skill la mejora.
7. **Una línea de base sin la skill**, medida una vez por una persona, fuera de un run: las mismas preguntas con el
   mismo modelo y sin kitlegal. Va a `evidencias/` y a `docs/USO.md`. Sirve para decir qué aporta la skill, no para
   decidir nada.
8. **Mejorar la skill sigue siendo trabajo normal**, que se mide con la serie y no bloquea otros hitos. El
   primer candidato es dar la rúbrica de cada artículo en el índice: sola no baja las marcas, pero en el sondeo del
   2026-10-11 el juez acepta 17 de las 19 menciones que la usan.

## Validación

Probabilidad de que una ejecución de 108 respuestas pase, según la tasa real de la skill:

| Umbral | Tasa de hoy (0,9 %) | 1,5 % | 3 % | La de H22 (5,6 %) | La de H21 (10 %) |
|---|---|---|---|---|---|
| 0 de 108 (hoy) | 37 % | 20 % | 4 % | 0,2 % | 0 % |
| ≤ 1 | 73 % | 52 % | 16 % | 1,5 % | 0 % |
| ≤ 2 | 92 % | 78 % | 37 % | 6 % | 0,1 % |
| **≤ 3** | **98 %** | 92 % | 59 % | **14 %** | 0,4 % |
| ≤ 4 | 99,6 % | 98 % | 78 % | 28 % | 1,4 % |

Con «como mucho 3», una ejecución con la tasa de hoy sale en rojo 1 de cada 50 veces, y una vuelta a la tasa de
H22 se detecta 6 de cada 7. Una regresión no se va con la ejecución siguiente: que dos seguidas no la vean pasa 1
de cada 50 veces. Con «como mucho 2» se detecta algo más (94 %), a cambio de un rojo sin motivo 1 de cada 12
ejecuciones, que es volver a un control del que no se puede fiar el rojo.

En `jurisprudencia`, con 60 respuestas, «como mucho 1» pasa el 96 % de las veces con una tasa del 0,5 % y detecta
una del 5 % 4 de cada 5 veces y una del 10 % casi siempre.

Las tres mediciones del cierre de H25 dieron 1, 2 y 2 de 108: las tres pasarían. Eso no las convierte en el cierre
de H25: una medición vale para el producto que midió (ADR 0030), y la #135 se mide de nuevo con este umbral en el
job.

## Consecuencias

**A favor**

- El verde y el rojo vuelven a decir algo: el job falla cuando la skill empeora, no cuando sale una respuesta suelta.
- Un hito deja de caer por una skill que no toca, sin dejar ningún umbral solo publicado.
- Desaparece el motivo para volver a medir hasta que salga limpio y para reparar a ciegas en el cierre.
- La tasa se mide y se enseña, con su total, en lugar de un cero que no era cierto.

**En contra, y asumido**

- **Se acepta que alrededor de 1 de cada 100 respuestas diga algo de un precepto que no ha leído.** Casi siempre es
  cierto y sin fuente; alguna vez estará mal.
- **Una subida pequeña no la ve ningún control.** Con 108 respuestas, una skill que pasara del 0,9 % al 3 % seguiría
  pasando 6 de cada 10 ejecuciones. La tasa acumulada lo enseña, pero solo se publica.
- **La regla «nunca inventar contenido legal» sigue en la skill como regla**, y lo que el proyecto dice hacia fuera
  tiene que ser lo medido, no «nunca».
- **En `jurisprudencia` se afloja un umbral que hoy se cumple.** Una respuesta de 60 que resuma una sentencia que
  no ha leído ya no pone el job en rojo; dos, sí. Se hace para que su primera marca sea un dato de la serie y no
  otro cierre caído por sorteo.

## Lo que este ADR no decide

- La frontera de la rúbrica (opción 5).
- Qué modelo decide las evals (ADR 0016 y 0031).
- Si `reparar_cierre` puede tocar una skill por un umbral del juez. Con un umbral que casi nunca falla sin motivo
  importa menos, pero sigue siendo una reparación que el run no puede validar (ADR 0032).

## Enmiendas que provoca

Van en la propuesta que acepta este ADR:

- **ADR 0037**: su estado dice que este lo sustituye en el umbral del punto 7.
- **`evidencias/adr-0038/`**: el material de los cuatro sondeos.
- **`docs/USO.md`**: la entrada del 2026-10-12.
- **`docs/WORKFLOW.md`**: la fila de la evidencia de un ADR en su tabla de rutas.
- **`CLAUDE.md`**: una decisión cerrada más.

Van con el hito que lo implemente, cuya entrada en `docs/ROADMAP.md` escribe Jorge:

- **`evals/boe-legislacion/juez/clases.yaml`** y su copia de `evidencias/adr-0037/`, y los de `jurisprudencia`
  en la #135: el umbral. La medida del juez no se repite: depende de la rúbrica, los casos, el modelo y la versión,
  y ninguno cambia.
- **`internal/evals`**: el umbral de la clase sobre los dos modos juntos, y el recuento por modo publicado.
- **La serie de cada skill** y `scripts/workflow/informe.sh`, que la enseña con la tasa acumulada.
- **`docs/ROADMAP.md`** (las secciones de H24 y de H25), **`CONTRIBUTING.md`** y lo que `docs/WORKFLOW.md` y
  `CLAUDE.md` dicen del job: el umbral.
