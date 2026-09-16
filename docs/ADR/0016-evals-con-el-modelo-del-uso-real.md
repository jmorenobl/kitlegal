# 0016 · Las evals deciden con el modelo del uso real, por repeticiones y cuando hay algo que medir

- **Estado**: aceptada
- **Fecha**: 2026-09-16
- **Hito**: pieza aparte tras H5; enmienda la clarificación Q5, FR-070 y SC-003 del spec de H5, su research D13 y la
  tabla de controles de `docs/ROADMAP.md` §4, sin tocar el contrato `Applet` (ADR 0005) ni el binario distribuido

## Contexto y problema

H5 dejó el job de evals (`.github/workflows/evals.yml`, `scripts/evals.sh`, `internal/evals`) decidiendo con
`claude-haiku-4-5-20251001`, una sola sesión por eval, y ejecutándose además una vez por semana sobre `main`
(`cron '41 4 * * 1'`). Eso viene de la clarificación Q5 del spec —«un único modelo de gama económica»— y esa, a su vez,
de la fila «job semanal con modelo barato» de la tabla de controles del roadmap.

Al cerrar el hito (bitácora `docs/USO.md`, entradas del 2026-09-15 y del 2026-09-16) aparecieron cuatro cosas que esa
letra no resuelve:

- **El modelo que decide no es el del uso real.** La skill `boe-legislacion` la usa quien la usa con el modelo que
  tenga; medirla con el más barato de la gama dice si la skill funciona en el peor caso, no si funciona. Tres tareas de
  H5 existen solo para acomodar lo que ese modelo hace de forma no fiable: T041 y T043 quitaron las preguntas por
  materia y pusieron el número del artículo en las diez positivas, y T046 tuvo que admitir la forma legible dentro de
  los corchetes de la cita porque, tras dos refuerzos del protocolo, el modelo seguía metiéndola ahí en una sesión de
  cada diez.
- **«10 de 10 en una sola tirada» es frágil con cualquier LLM.** Las dos ejecuciones de cierre (35002104338 y
  35023013878) dieron 10 de 10, pero entre las ejecuciones 34941499481 y 34956596912 tres evals cambiaron de resultado
  sin que cambiara nada del repositorio. Un criterio que una repetición puede volcar no mide la skill: mide la tirada.
- **Las preguntas por materia salieron del conjunto y no dejaron ningún rastro mecánico.** Quedaron solo en la
  bitácora, y la herramienta que las haría posibles —encontrar dentro de una norma el artículo que trata una materia—
  sigue en el backlog. Sin ellas en `evals/`, nada avisa de que la skill cubre menos de lo que un usuario le pide.
- **La ejecución semanal no mide ningún cambio.** Sobre `main`, con el modelo, la versión de Claude Code y las
  respuestas del BOE fijados, lo único que varía de una semana a otra es el azar del modelo. A cambio, gasta
  suscripción y asume cada semana el riesgo del token legible en `/proc/<pid>/environ` que documenta research D13 de H5.

Hay además una premisa que conviene dejar escrita, porque es la que más veces se razona mal: **Jorge no usa la API de
pago por uso.** Las sesiones del job consumen los límites de una suscripción de Claude, con el secreto
`CLAUDE_CODE_OAUTH_TOKEN` que genera `claude setup-token`. Un modelo más caro por token no cuesta más dinero aquí: lo
que consume es cuota. Así que «modelo barato» nunca fue un criterio de calidad, y el roadmap lo escribió como si lo
fuera.

## Opciones consideradas

### Qué modelo decide

1. **Seguir con Haiku 4.5.** Rechazada: mide el peor caso y ya ha obligado a rebajar tres veces lo que se pide.
2. **Decidir con Opus 5.** Rechazada: si funciona con Sonnet funciona con Opus, así que Opus no añade información y
   consume más cuota de la suscripción; y la mayoría del uso real no va con Opus.
3. **Decidir con Sonnet 5** (`claude-sonnet-5`), y seguir ejecutando Haiku 4.5 como límite inferior publicado en el
   informe, sin que haga fallar el veredicto. Así el veredicto habla del uso real y el informe conserva, gratis, la
   respuesta a «¿y con el modelo más barato?».

### Cómo se decide una eval

1. **Una sesión por eval, verde exigido.** Es lo que hay, y es lo que la bitácora documenta como frágil.
2. **Una sesión por eval, verde exigido, con reintento si falla.** Rechazada: un reintento que solo se pide cuando
   falla convierte cualquier tasa en «pasa si alguna vez pasa», que es el criterio más flojo posible y además no se
   publica.
3. **N sesiones por eval y umbral**, con la tasa en el informe. Con N = 3 y umbral 2, una eval pasa con 2 de 3. La tasa
   se publica siempre, también la de las series que pasan, de modo que una degradación se ve antes de volverse roja.

### Las preguntas por materia

1. **Fuera del conjunto, solo en la bitácora.** Es lo que hay: nada mecánico avisa.
2. **Dentro del conjunto y decidiendo.** Rechazada: harían fallar el job por una herramienta que aún no existe, y
   romperían SC-003 sin que nadie pueda arreglarlo hasta que llegue esa herramienta del backlog.
3. **Un directorio aparte, `evals/<skill>-materia/`.** Rechazada: duplicaría las reglas del conjunto, la preparación de
   la caché y el informe, y el valor de estas preguntas está justamente en compararlas, eval a eval, con la que sí
   nombra el artículo.
4. **Dentro del conjunto, marcadas `informativa: true`**: se ejecutan, su tasa se publica y no deciden el veredicto.
   Una regla del conjunto exige que haya al menos una y que todas sean positivas, de modo que no pueden volver a
   desaparecer en silencio.

### Cuándo se ejecuta

1. **Semanal sobre `main`.** Rechazada por lo de arriba.
2. **Semanal más por cambios.** Rechazada: la semanal seguiría sin medir nada y sumaría su coste al nuevo.
3. **Por cambios, a mano y por etiqueta.** Arranca al abrir o reabrir una propuesta de cambio que toque lo que las
   evals miden, y siempre que se ponga la etiqueta `evals` o `evals-prueba-de-red`. Sin `synchronize`: cada ejecución
   abre decenas de sesiones con modelo, y un hito empuja muchas veces; la etiqueta es el botón explícito de «vuelve a
   medir», y es la que usa la ejecución de cierre de un hito (FR-082).

## Decisión

1. **Decide `claude-sonnet-5`**, fijado por su id en `MODELO_DE_EVALS`. **`claude-haiku-4-5-20251001` sigue
   ejecutándose** en `MODELOS_INFORMATIVOS_DE_EVALS`: su tasa se publica como límite inferior y no hace fallar el
   veredicto. Un modelo informativo no abre las evals informativas.
2. **Cada eval se repite** `REPETICIONES_DE_EVALS` veces con cada modelo (3) y su serie pasa con `UMBRAL_DE_EVALS`
   (2). El informe publica la tasa de cada serie, también las que pasan.
3. **Vuelven las preguntas por materia** como cinco evals `informativa: true` (13 a 17 de `evals/boe-legislacion/`),
   con la redacción literal que registró la bitácora el 2026-09-15 para poder compararlas con lo que allí se midió.
4. **La extracción tolerante de la cita se queda** como la dejó T046: compara por identificador, que es lo que SC-009
   quiere medir, y la forma legible dentro de los corchetes no cambia qué norma y qué bloque se citan.
5. **El disparador es por cambios**, a mano y por etiqueta; se quita el `schedule`.

Lo que **no** cambia: la garantía de red (FR-074, FR-076), la retirada de Python (FR-073, FR-081), el tope de 240 s por
sesión —atado a la vigencia de 300 s de `buscar` y `metadatos` en la caché—, la lectura de la traza, y que el job siga
fuera de `make ci`.

## Consecuencias

**A favor.** El veredicto habla del modelo con el que se usa la skill. Una eval que falla una vez de tres deja de
volcar una ejecución entera y pasa a verse como una tasa. Las preguntas por materia vuelven a estar en el repositorio,
medidas, sin poder tumbar el job. El job se ejecuta cuando hay algo que medir, y la ejecución semanal deja de asumir
cada lunes el riesgo del token.

**En contra, y asumido.**

- **Cada ejecución cuesta mucho más.** Con 17 evals, 3 repeticiones y dos modelos, una ejecución abre 87 sesiones
  (36 de Sonnet sobre las 12 que deciden, 15 de Sonnet sobre las 5 informativas y 36 de Haiku sobre las 12 que
  deciden), frente a las 12 de antes. El presupuesto: la preparación del runner midió 5 min en la ejecución 35002104338
  (193 s solo la retirada de Python), así que dentro de `timeout-minutes: 120` quedan unos 6 900 s para sesiones y la
  media por sesión tiene que quedar por debajo de ~79 s. Haiku midió ~21 s por sesión, preparación incluida; de Sonnet
  no había medida al decidir esto, y la primera ejecución de la propuesta de cambio es la que la da. Si no cabe, la
  salida no es volver a un modelo peor: es subir `timeout-minutes` o bajar las repeticiones, y decirlo.
- **Un fallo aislado ya no se ve en el veredicto.** Se ve en la tasa y en la tabla de sesiones del informe, que hay que
  mirar. A cambio, lo que el veredicto dice es más fiable.
- **Las evals informativas no protegen de nada.** Miden y publican; que la 13 a la 17 estén en rojo es el estado
  esperado hasta que llegue del backlog la herramienta que busca un artículo por su materia. La regla del conjunto
  garantiza que sigan ahí, no que pasen.
- **Una propuesta de cambio que solo toca `docs/` no arranca el job.** Es deliberado; la etiqueta lo lanza igual.
- **`gh api` en el job de filtro** añade un permiso (`pull-requests: read`) y una dependencia de la API para decidir si
  hay algo que medir. Se prefiere a un filtro `paths:` del evento porque `paths:` se aplica también a la actividad
  `labeled` y dejaría sin arrancar la etiqueta sobre una propuesta que no toca esas rutas, que es justo el lanzamiento
  que FR-070 conserva.

## Enmiendas que provoca

Marcadas como enmienda con su fecha y su motivo, sin reescribir lo que decía antes:
`specs/006-h5-skill-boe-legislacion/spec.md` (clarificación Q5, FR-070 y SC-003), su `research.md` (D13), su
`data-model.md` (§6, §6.3, §10.2, §10.3 y el nuevo §10.4), su `contracts/job-de-evals.md` y su `quickstart.md` §12.3; y
`docs/ROADMAP.md` §4, cuya fila de controles de skills decía «job semanal con modelo barato».
