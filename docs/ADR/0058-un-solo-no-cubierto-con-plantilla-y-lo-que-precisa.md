# 0058 · Un solo `⚠ NO CUBIERTO:` con plantilla, y lo que precisa de los ADR 0044 y 0048

- **Estado**: propuesta (2026-11-18); queda aceptada cuando una persona fusiona la propuesta de cambio que la trae.
- **Fecha**: 2026-11-18
- **Hito**: H33. Se escribe con su entrada (`docs/roadmap/todo/050-H33-cobertura-en-positivo-por-capacidad.md`, #277),
  antes del run y fuera de él, en una sesión interactiva. Decide en siete puntos lo que la entrada necesita y que el ADR
  0044 (punto 7) pide decidir con un ADR, en lugar de dejarlo a la entrada.

## Contexto y problema

El ADR 0044, punto 4, pide que H33 dé a todas las materias un único `⚠ NO CUBIERTO:` que diga también con qué
herramienta busca el agente lo que falta. Al escribir la entrada, y en dos revisiones independientes con un prototipo
fuera del repositorio, salieron cinco puntos en los que la entrada no puede seguir lo escrito sin reabrir otra decisión
o sin dejar el cierre en riesgo:

- **Nombrar la herramienta en la línea.** La línea se repite en las nueve filas de lo regulado aparte. Tras ellas no se
  expone la norma que rige el procedimiento (ADR 0043, punto 7, que nació de una sesión que leyó los arts. 223 y 235 LGT
  porque el texto no lo prohibía). Una cola que dijera «busca con…» o «lo que lea de su norma va marcado» invitaría a
  esa lectura en cada fila. Y la herramienta de búsqueda, `boe buscar`, no la sirven las evals con texto libre (ADR
  0043, puntos 2 y 3).
- **«En prosa»** (ADR 0048, punto 1): `consulta-juridica` dice lo no cubierto con sus palabras. El ADR 0044 lo cambia,
  pero es anterior y no dice cómo.
- **El molde del paquete** (ADR 0048, punto 5) es «sus normas, una referencia por skill que la necesita y lo que la ley
  regula aparte, y nada más». Con lo cubierto en positivo, un paquete trae también lo que cubre y los motivos de su
  línea.
- **El territorio.** El ADR 0044, punto 4, lo nombra dentro de lo administrativo. `consulta-juridica` lo declara cubierto
  para cualquier pregunta desde H32, sin materia. Y la identificación de un municipio no depende de ninguna (principio
  IX).
- **`sin_activar` en `consulta-juridica`.** Con una eval de cobertura, el código de `internal/evals` da a la skill sus
  umbrales de respuestas, y entre ellos `sin_activar`, con 0 de tope. Hoy no lo tiene, y la serie es corta: en el job de la
  v0.10.1, las 12 sesiones del modelo que decide la activaron, y en el cierre de H41.4, sus 12 también.
- **Dos rutas válidas para una pregunta.** Al medir el prototipo antes de lanzar salió el séptimo. Con lo cubierto de
  los plazos completo —el acceso a la información, que `data/plazos.yaml` calcula—, la eval 17 de `normativa` («¿En qué
  plazo hay que resolver una solicitud de acceso…?») la resuelve también `plazos-y-territorio`, que lee el precepto con
  su paso 6. En un sondeo de 240 sesiones, 28 respondieron solo con ella, las 28 bien y citadas, y `sin_activar` de
  `normativa` las contaba como sin activar. Con una sola skill instalada, «activó la skill» era «usó el producto»; con
  cuatro (ADR 0048), puede ser elegir entre dos rutas que el producto permite.

## Decisión

1. **La línea es una plantilla, y su cola no nombra ninguna herramienta.** La línea es `⚠ NO CUBIERTO: <asunto>:
   <motivo>. kitlegal no da sus plazos, órganos, vías ni escritos.`, la misma en todas las materias y en toda skill que
   la lleve. El `<motivo>` dice por qué no está cubierto: fuera de lo declarado, un texto fijo de `data/cobertura.yaml`,
   sin verbo que concordar con el asunto, porque se comprueba palabra por palabra; en lo regulado aparte, el precepto
   que lo dice y, si se cita de lo grabado, la norma que lo rige. Qué se puede leer para completar la respuesta lo dice
   la skill, no la línea. Sustituye, en lo que dice de la línea, la última viñeta del punto 4 del ADR 0044.
2. **Lo leído para completar va marcado, y solo fuera de lo declarado.** `consulta-juridica` lleva la marca
   `⚠ NO VERIFICADO:` para lo que el agente lea con la skill de normativa tras su línea. El texto citado es el de la
   norma, y lo no verificado es que se aplique al asunto. En lo regulado aparte sigue el ADR 0043, punto 7: no se expone
   la norma que lo rige y no hay marca. `plazos-y-territorio` no lleva la marca: no lee normas de otra materia. Exponer
   marcada la norma de lo regulado aparte pediría un ADR que sustituya ese punto.
3. **`consulta-juridica` declara lo no cubierto con la línea**, no en prosa. Sustituye el punto 1 del ADR 0048 en eso;
   lo demás de ese punto sigue.
4. **Un paquete de materia trae también sus filas de lo cubierto**, en `data/cobertura.yaml`, y, si trae la región de
   lo regulado aparte, los motivos de su línea en esa referencia. Precisa el punto 5 del ADR 0048. Sigue sin código y
   sin cambiar ningún protocolo fuera de lo que genera la sincronía.
5. **El territorio está cubierto en todas las materias**, como lo declara `consulta-juridica` desde H32. La provincia,
   la comunidad, el régimen, el DIR3 del ayuntamiento y los boletines de un municipio salen de registros nacionales y no
   dependen de la materia (principio IX). Ceñirlos a lo administrativo quitaría cobertura en uso, y H33 no quita ni
   añade cobertura. Precisa el punto 4 del ADR 0044.
6. **`sin_activar` nace en `consulta-juridica`, con 0 de tope**, como en las demás skills con umbrales de respuestas:
   activarse es un hecho de la sesión (ADR 0038, punto 1). No se le hace una excepción en el código, que sería un caso
   especial por skill. El riesgo de un cierre en rojo por un umbral sin serie lo reduce, sin quitarlo, la medida de antes
   de lanzar que fija la entrada: una ejecución completa del job sobre el prototipo, y un rojo se compara con `main`.
   Con 12 respuestas por modo en `sin_activar`, un 0 no descarta una tasa de fondo baja; si el cierre cae por él, es la
   serie que ve la persona (ADR 0038, punto 5), no un defecto que el run repare.
7. **Una eval puede declarar otras skills que también la resuelven** (`tambien_la_resuelven`). Su activación cuenta
   como la de la skill de la eval, en el juicio y en `sin_activar`, cuyo tope sigue en 0: lo que sigue siendo un
   defecto es responder sin ninguna skill del producto que la resuelva. Lo declara una persona en el fichero de la
   eval, con un comentario que dice por qué; cada nombre es una skill de `skills/`. No cambia lo que la sesión tiene que
   hacer —sus comandos y sus citas— ni ningún tope, y la ruta que tomó cada sesión sigue en el informe. Lo lleva la eval
   17 de `normativa`, con `plazos-y-territorio`. Es un cambio del criterio hecho tras ver los datos, y por eso se dice:
   la ruta la permite el paso 6 de `plazos-y-territorio`, y las respuestas eran correctas. Precisa el punto 6 del ADR
   0038 y el ADR 0054 (opción 1), que rechazó dar un margen a `sin_activar`: esto no es un margen, sino qué skill
   cuenta.

## Consecuencias

- **La cola es más corta.** El cuerpo de `plazos-y-territorio` baja de 10 882 a 10 688 caracteres, con sitio para la
  línea de cuatro paquetes, contando el de ensayo de los tests.
- **Las nueve filas de lo regulado aparte dejan de decir «no calcula sus plazos ni sus recursos»**, y pasan a decir que
  kitlegal no da sus plazos, órganos, vías ni escritos. Lo que hace la skill tras la línea no cambia.
- **La primera medida de la plantilla y del `sin_activar` de `consulta-juridica` la hace la persona antes de lanzar**,
  no el cierre del run.
- **La regla «frontera en dos materias»** pide dos normas sin `vertical`. El paquete que tome la Ley de Contrato de Seguro
  o el Estatuto de los Trabajadores tendrá que mover la eval de ese lado a otra norma sin paquete.
- **`sin_activar` mide si se usó el producto, no qué skill eligió el modelo**, en las evals que lo declaran. Una
  pregunta con dos rutas ya no tumba un cierre por la ruta.
- **Lo que falta medir sigue en la bitácora.** Una línea de no cubierto de más en una pregunta cubierta, y una fecha
  contada a mano en la prosa de una respuesta que lleva la línea, no los decide ningún umbral. La entrada los deja como
  límites conocidos.

## Validación

- **Los textos, los caracteres y las suites** se comprobaron en un prototipo fuera del repositorio, sobre `3963b85e`.
  Llevaba la plantilla, el generador, la región de lo cubierto, el formato de eval, las evals 01, 03 y 04 y la vertical
  de ensayo con su fila y sus motivos. Con `go test ./...` y `-tags=integration`, solo cayeron los tests que la entrada
  enumera.
- **`sin_activar` en `consulta-juridica`** salió en el prototipo con `Decide: true` y 0 de tope en los dos modos
  (`TestUmbralesDeLasSkills`).
