# 0043 · Una eval no mide la avería de su propia grabación, salvo que la provoque

- **Estado**: propuesta
- **Fecha**: 2026-10-20
- **Hito**: H26.1. No toca el punto 7 del ADR 0018: lo que se graba lo sigue grabando `grabar_datos` tras el commit
  de una tarea `[datos]`, solo de fuentes con fila revisada en `docs/SOURCES.md` de `main`. Cambia qué sirve la caché
  preparada de una sesión de evals (data-model §7.1 de H5, ampliado en H20 con la lectura a `9999-12-31`) y cumple el
  FR-052 de H9, que pedía que lo que leen las sesiones de `legal-core` estuviera grabado para el job. No cambia el
  contrato `Applet` (ADR 0005), ni el sobre (ADR 0006), ni la tabla de resultados (ADR 0023), ni el umbral de ninguna
  clase del juez (ADR 0038), ni el modelo que decide (ADR 0031).

## Contexto y problema

Las sesiones de evals corren sin red. Cada una arranca con una caché preparada que solo sirve las consultas que nombran
las evals de su skill —sus comandos esperados, sus citas y el índice y los metadatos de sus normas
(`ConsultasNecesarias`, `internal/evals/consultas.go`)—, no todo lo grabado. Lo que la caché no sirve sale con 5,
porque la sesión no alcanza el `robots.txt` del BOE, como si el BOE estuviera caído. En el uso real esas lecturas
funcionan: la eval mide una avería que no se da.

En los jobs de evals de la #175 (run 39447467111) y de la #178 (run 39492300808):

- **`legal-core`.** Una sesión de `claude-sonnet-5-5` en modo herramienta de la eval 07 (una liquidación de la Agencia
  Tributaria) leyó la DA 1ª de la LPAC (`BOE-A-2015-10565`, bloque `daprimera`), que salió con 5, y respondió sin la
  línea `⚠ NO CUBIERTO:`: `cobertura_sin_declarar:claude-sonnet-5-5:herramienta`, que decide con 0, dio 1 de 6. El
  bloque **está grabado**: lo grabó H9 en el manifiesto de `internal/evals/testdata/` para «lo que leen del BOE las
  sesiones de las evals de `legal-core`». Ninguna eval lo nombra, así que la caché no lo sirve. Lo mismo pasó con el
  art. 30 LPAC leído a la fecha de inicio en la eval 05, y con el índice del texto refundido de la Ley sobre Tráfico
  (`BOE-A-2015-11722`) en la 08, también grabados. `legal-core` v1.2.2 (#178) hizo que la respuesta aguante la lectura
  fallida, pero el hueco sigue.
- **`boe-legislacion`.** `afirma_lo_no_leido:claude-sonnet-5-5`, que decide con una tasa de 0,028 sobre las respuestas
  de los dos modos, dio 3 de 120 (cabían 3,36). Una de las tres viene del hueco: en la eval 16 (el principio de
  legalidad de la potestad sancionadora, por materia), la sesión pidió los arts. 26 y 27 de la Ley 40/2015, que están
  junto al 25 en el mismo capítulo; no están grabados, salieron con 5, y la respuesta dijo de qué trataban sin
  haberlos leído. Las mismas dos lecturas fallaron en 4 sesiones de las dos tandas, y el art. 18 de la Ley 35/2006
  (eval 15) y el 15 del texto refundido de la Ley reguladora de las Haciendas Locales (eval 14), en una cada uno.
- **`jurisprudencia`.** Ninguna lectura fuera de lo grabado: `cita` no consulta ninguna fuente.
- **Las demás lecturas que fallaron** son de `claude-haiku-4-5`, que en `legal-core` no activó la skill en ninguna
  sesión de las evals 05 a 10 de la #175, o son lo que la skill dice que no se haga: buscar normas cuyo identificador
  ya está en las referencias, aplicar la alzada a una multa de tráfico (regla 9 de `legal-core`) o leer la Ley 30/1992
  para un plazo de 2014 que `plazos` no calcula (paso 9). Grabarlas haría pasar mejor una respuesta peor.

Y el caso contrario: si la caché sirve `daprimera`, nada comprueba ya lo que arregló la #178, que `legal-core` promete en
el paso 7 y en la regla 9. Hoy ninguna eval puede hacer fallar una lectura a propósito: el formato
(`schemas/eval.yaml.json`) no tiene ningún campo para eso. La única avería que provoca es la falta de binario y de
servidor (`sin_binario_ni_servidor`).

## Opciones consideradas

1. **Grabar más y seguir sirviendo solo lo que nombran las evals.** Rechazada: `daprimera` ya estaba grabado. El
   defecto no está en lo que se graba sino en lo que se sirve, y cada lectura nueva pediría añadir a una eval un
   comando o una cita que la respuesta no tiene por qué hacer.
2. **Servir todo lo grabado.** Rechazada. Las búsquedas se guardan con una huella de su consulta y no se pueden volver a
   convertir en consultas; servir lo grabado durante la sesión, en lugar de preparar su caché, pediría al binario un
   modo de reproducción que el binario publicado no tiene, y la sesión dejaría de medir ese binario. Además, lo grabado
   incluye respuestas de error de H4 que ninguna skill pide.
3. **Una lista de lecturas por eval, y nada más.** Rechazada como único mecanismo: repetiría en cada eval los preceptos
   que ya están en `data/`, y cada hito que añada un tipo de plazo tendría que acordarse de añadirlos.
4. **Servir lo que la skill lleva a leer, derivado de `data/`; una lista por eval solo para la exploración razonable;
   y un campo que haga fallar una lectura a propósito.**

## Decisión

Opción 4.

1. **La caché preparada sirve lo que lleva a leer el texto de las skills, derivado de `data/`**, además de las consultas
   necesarias de las evals: cada bloque de `data/plazos.yaml` (los de los tipos, de las unidades y de los cómputos) y
   de `data/cobertura.yaml` (el precepto de cada entrada y el aplicable de su norma), sin fecha y con
   `--fecha 9999-12-31`, como los demás bloques; y el índice y los metadatos de cada norma de esos dos ficheros y de
   `data/normas.yaml`. Es la clase (a) de lo que lee una sesión: los preceptos que la skill nombra en sus referencias y
   en sus reglas. Vale para las sesiones de todas las skills, que tienen todas instaladas. Todo ello ya está grabado
   (comprobado el 2026-10-20: 17 bloques, y el índice y los metadatos de cada norma); y `TestPreceptosDeLosPlazos` ya
   exige que los bloques de los dos ficheros estén grabados. Si un hito añade un precepto a `data/` sin grabarlo,
   `Preparar` da la falta y `make ci` falla.
2. **La exploración razonable va en la eval que la provoca.** Una eval puede llevar `lecturas`: bloques que la caché de
   la skill sirve, cada uno con su motivo, sin que la respuesta tenga que leerlos y sin comprobar nada de ellos. Es la
   clase (b): la lectura que haría una respuesta mejor, de la norma de la pregunta, vista al menos en una sesión del
   modelo que decide con la skill activada. Se graba como todo, con `grabar_datos` en una tarea `[datos]` de un hito.
   No entran nunca en `lecturas`: una búsqueda de texto libre, que cambia con cada redacción; lo que lee una sesión sin
   la skill activada; y lo que la skill dice que no se haga.
3. **Lo que la skill no pide ni se graba ni se sirve.** Es la clase (c). Si la lectura apunta a un hueco del texto de la
   skill, lo que se corrige es la skill.
4. **Una eval puede hacer fallar una lectura a propósito** con `sin_servir`: los bloques, o el índice o los metadatos
   de una norma, que la caché de sus sesiones no sirve aunque los necesite otra eval o los traiga `data/`, tampoco con
   `--fecha`. La lectura sale con 5, como con el BOE inalcanzable, que es la avería real. Un bloque de `sin_servir` no
   puede ser una de sus `citas`. El informe dice, por eval, cuántas sesiones intentaron la lectura que no se sirve:
   una eval cuyas sesiones nunca la intentan no comprueba nada, y eso se tiene que ver.
5. **El informe del job cuenta las lecturas fuera de lo grabado** con `fuera_de_lo_grabado:<modelo>:<modo>`, sin las
   que provoca `sin_servir`, publicado y sin decidir: un hueco nuevo se ve en el resumen sin buscarlo en el registro.
6. **Si falla la lectura de un precepto a la fecha de inicio, la línea del plazo se mantiene.** La fecha y su precepto
   salen de `plazos`, que no consulta la red y que ya comprobó contra el BOE grabado que ese apartado regía el día de
   inicio (ADR 0042); con el BOE caído la persona sigue necesitando la fecha. La respuesta da `⚠ PLAZO:` con la cita de
   cada precepto de `data.preceptos`; en lugar de `⚠ REDACCIÓN A FECHA:`, dice aparte que no pudo leer esa redacción,
   como `boe-legislacion` dice cualquier lectura fallida, y no afirma lo que dice el precepto. La comprobación de citas
   del job admite en la línea `⚠ PLAZO:` el precepto que da el resultado de `plazos` de la sesión.
7. **Tras `⚠ NO CUBIERTO:`, `legal-core` no expone la norma que rige el procedimiento.** La nombra la línea, con lo que
   trae `references/cobertura_por_materia.md`; sus plazos y sus recursos llegan con su vertical (ADR 0027). Una sesión
   leyó los arts. 223 y 235 de la Ley General Tributaria en la eval 07: el texto de la skill no lo prohibía.

## Consecuencias

- Las lecturas que la skill pide dejan de fallar en las evals por no estar en la caché, y lo que añada un hito a
  `data/` se sirve solo. Las lecturas de la clase (b) se suman por eval, con su motivo, y se graban dentro de un hito.
- La robustez ante una lectura fallida pasa a medirse a propósito, con las evals de `sin_servir`, en lugar de por
  azar.
- Ninguna sesión gana red ni lee otra cosa que lo grabado: el binario de la sesión sigue siendo el publicado.
- Grabar sigue pidiendo un hito: fuera de un run no hay tarea `[datos]` ni `grabar_datos`, y este ADR no lo cambia.
- Siguen aparte, porque no son de grabación: que `claude-haiku-4-5` casi no active `legal-core` en el job, y el código 1
  (defecto inesperado) de `boe indice` con los índices grandes que se ve en el job de `boe-legislacion`.
