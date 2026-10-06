# 0036 · Del CENDOJ, resolver una resolución identificada por su formulario; la búsqueda la hace la persona

- **Estado**: aceptada (2026-10-03, #109). Enmendada el 2026-10-06, antes de lanzar H23, sin cambiar la decisión
  («Enmienda: lo que la entrada de H23 contradecía»).
- **Fecha**: 2026-10-02
- **Hito**: transversal (tras H22, antes del hito de jurisprudencia del backlog «fuentes»). Sustituye del ADR
  0003 el primer punto de su decisión (resolver por el resolutor ECLI europeo, «no del buscador») y corrige dos
  hechos de su contexto. El resto del ADR 0003 sigue vigente: ni ingesta masiva, ni espejo, ni texto íntegro
  descargado automáticamente, ni resumen de lo que no se ha leído.

## Contexto y problema

El ADR 0003 decidió en H0 que del CENDOJ solo se resuelven ECLI y metadatos, y que se resuelven por el resolutor
ECLI del portal e-Justice, nunca por el buscador. Tres cosas que entonces se dieron por ciertas no lo son:

1. **El resolutor europeo no resuelve.** Por HTTP devuelve un armazón vacío y su servicio SOAP está declarado no
   disponible (`docs/JURISPRUDENCIA.md` §2, 2026-09-13). La decisión del ADR 0003 no se puede implementar para
   ningún órgano que no sea el Tribunal Constitucional: sin el formulario no hay forma de saber si una sentencia
   del Supremo existe.
2. **El aviso legal no prohíbe la consulta automatizada.** Su texto limita el uso al particular y prohíbe el uso
   comercial, la descarga masiva y la elaboración de bases de datos fuera del procedimiento del CGPJ. El ADR
   0003 dice que los términos «prohíben expresamente el uso masivo o automatizado»; lo segundo no está en el
   aviso. `robots.txt` no excluye `/search/` y pide cinco segundos entre peticiones.
3. **Una consulta no encuentra CAPTCHA.** El 2026-10-02, una consulta por el campo `ECLI` del formulario, por
   HTTP simple y con el agente identificable de kitlegal, devolvió la resolución con sus metadatos y su URL, sin
   navegador, sin JavaScript y sin CAPTCHA (`docs/JURISPRUDENCIA.md` §3). La página carga una hoja de estilos de
   CAPTCHA, así que existe en algún punto que no se ha encontrado.

El ADR 0003 dice que no se reabre «por conveniencia ni por presión de una funcionalidad concreta», y lo que ha
traído la pregunta es eso: el benchmark de asistentes legales de observatorio.legal, donde la verificación de
citas de jurisprudencia es lo que más pesa. Este ADR no se apoya en esa presión, sino en que los tres hechos de
arriba eran inexactos y en que la decisión, tal como se escribió, no da lo que prometía: citas de sentencias
«resolubles y verificables». Queda dicho para que quien lo acepte sepa de dónde viene.

A favor de leer el aviso así: el reglamento del CGPJ que fijaba licencias y precios (3/2010) fue anulado por el
Tribunal Supremo en 2011 y no se ha localizado sustituto, y la Ley 37/2007 se aplica a las sentencias
[BOE-A-2007-19814, bloque `dasegunda`]. En contra: «uso particular» admite una lectura estrecha en la que el uso
de un despacho no cabe, y quien interpreta el aviso es el CGPJ, no kitlegal.

## Opciones consideradas

1. **Dejar el ADR 0003 como está.** kitlegal reconoce un ECLI del Supremo, da su ROJ y no puede decir si la
   sentencia existe. Coherente, pero la herramienta no comprueba justo lo que más daño hace cuando es falso.
2. **Buscar por materia y descargar los textos por el formulario**, como hace el guion de la skill fiscal
   anterior (un navegador sin ventana que se presenta como Chrome). Es descarga a escala repetida por cada
   persona y cada pregunta, se disfraza y depende de un navegador que el binario no lleva. Rechazada.
3. **Un modo desactivado por defecto, activado para medir.** Mide un producto que nadie tiene y no cambia lo que
   el aviso permite. Rechazada.
4. **Resolver por el formulario una resolución ya identificada, y dejar la búsqueda a la persona.**
5. **Pedir antes la licencia o la autorización al CGPJ y no hacer nada hasta tenerla.** No se sabe con qué norma
   se concede hoy ni en qué plazo, y no condiciona la opción 4.

## Decisión

Se adopta la **opción 4**.

- **Qué se consulta.** Una resolución que la persona o el modelo ya han identificado: por su ECLI, por su ROJ o
  por su órgano, número de resolución o de recurso y fecha. Se usa el campo del formulario que corresponde a ese
  dato. Una consulta por resolución, que nace de una pregunta de la persona en su equipo.
- **Qué se obtiene.** Si existe, sus metadatos (órgano, sala, fecha, número de resolución y de recurso, ponente,
  ROJ, ECLI) y su URL oficial. Cero resultados es «no encontrado» (ADR 0023) y la cita no se sostiene.
- **Qué no se hace.** No se busca por materia ni por texto libre. No se descarga, no se lee ni se guarda el
  texto de la resolución. No se resume ni se caracteriza una sentencia que no se ha leído. No se usa un navegador,
  no se cambia el agente por el de un navegador y no se sortea un CAPTCHA ni un bloqueo: si aparecen, el applet
  sale con la fila de «rate-limited/TOS» del ADR 0023.
- **Cómo se pide.** Solo por `internal/httpx`, con el agente identificable y el ritmo del `robots.txt` (5 s). Las
  evals y la integración continua no consultan el CENDOJ: reproducen grabaciones.
- **La búsqueda por materia** la hace la persona en su navegador con la consulta que le prepara la skill, y
  vuelve con el identificador, que se resuelve, o con el texto, que el agente lee (ADR 0004).
- **Antes de la primera release que lo lleve**, una persona revisa la fila de la fuente en `docs/SOURCES.md`, y
  la página `/bot/` de la web, a la que apunta el agente, dice qué hace kitlegal con el CENDOJ, con qué agente y a
  qué ritmo.
- **El Tribunal Constitucional** sigue por el BOE, como decía el ADR 0003.

## Consecuencias

**A favor**

- Una sentencia inventada deja de pasar: no se resuelve y no se cita. Es lo que el ADR 0003 quería y no daba.
- Lo que kitlegal hace con el CENDOJ es lo que hace una persona que comprueba una referencia, una vez, y se
  identifica al hacerlo.
- No entra ninguna dependencia nueva en el binario.

**En contra, y asumido**

- **El CGPJ puede no compartir la lectura** y bloquear el agente. La función dejaría de responder y kitlegal lo
  diría como fuente no disponible; no se buscaría otra forma de entrar.
- **La caché y el grafo guardan lo que ven.** Metadatos de las resoluciones que cada persona ha comprobado, en
  su equipo; el hito tiene que fijar qué se guarda y cuánto tiempo, y si el resumen del CENDOJ entra
  (`docs/JURISPRUDENCIA.md` §6).
- **Las grabaciones de los tests** llevan la página de resultados de unas pocas resoluciones en un repositorio
  público; se graban las mínimas.
- **No encuentra la sentencia aplicable** ni comprueba que diga lo que se le atribuye: para eso hace falta el
  texto, que trae la persona.
- El README («¿Y las sentencias?») y la web describen el ADR 0003 y cambian con el hito que lo implemente, no
  antes. La línea del CENDOJ en `CLAUDE.md` cambia con el roadmap que numera ese hito, porque guía a las sesiones
  de su run.
- Buscar por materia o leer textos de forma automática sigue pidiendo el procedimiento de reutilización del CGPJ
  y un ADR nuevo, no una excepción en el código.

## Enmienda: lo que la entrada de H23 contradecía (2026-10-06)

Hecha al comprobar la entrada de H23 antes de lanzarlo, y decidida por Jorge. No cambia la decisión: la lleva a
donde este ADR no llegó, y corrige la entrada del hito donde chocaba con él.

- **La constitución no se enmendó con este ADR.** Su principio I decía «No existe ninguna llamada HTTP con método
  distinto de GET/HEAD en el módulo» y «CENDOJ: solo resolución de ECLI y metadatos». El formulario del buscador
  se envía con `POST`, con la cookie de sesión de la página, y se consulta también por ROJ y por número de
  resolución con su fecha. `internal/httpx` rechaza hoy todo lo que no sea GET o HEAD y no guarda cookies, y los
  dos jueces de la revisión final del workflow lo exigen. Con la constitución como estaba, el hito no se podía
  especificar. **Constitución 2.12.0**: GET y HEAD, con una sola excepción, el envío del formulario de consulta
  de un buscador público cuya fila de `docs/SOURCES.md`, revisada por una persona, lo declara. Es una consulta
  sin identidad, que no presenta, firma ni cambia nada en la fuente: «nunca un POST a una sede» (ADR 0004) sigue
  como estaba. Workflow `hito` 2.5.1: la rúbrica de los jueces finales dice lo mismo.
  - Considerado y rechazado: probar si el buscador responde a un GET. El formulario es un `POST`, y usarlo como
    lo usa una persona es lo que sostiene la lectura del aviso legal.
- **La integración continua no consulta el CENDOJ, tampoco de noche.** La entrada de H23 ponía una consulta en el
  flujo nocturno, contra «Cómo se pide» y contra «una consulta por resolución, que nace de una pregunta de la
  persona». La comprobación de que la respuesta del buscador se sigue interpretando la lanza una persona, por su
  nombre, antes de una release o cuando alguien avisa. Lo que se pierde es enterarse solo de un cambio de forma:
  el applet lo dice a quien consulta, con la fila de «rate-limited/TOS».
- **Una cita probada como ROJ lleva su fecha.** La entrada decía que «STS 1088/2023, de 4 de julio» se prueba
  como número de resolución con su fecha y, si no da nada, como ROJ. Los ROJ son correlativos: casi cualquier
  número existe como ROJ de otra sentencia, y una cita inventada, o con la fecha mal recordada, quedaba
  comprobada con el enlace de un asunto ajeno. Con `--roj` y `--fecha`, lo encontrado con otra fecha no es lo
  pedido, y lo compara la herramienta, no el modelo. Una cita así sin fecha no se prueba como ROJ. Visto el mismo
  día, al traer el texto de abajo: «STS 1088/2023» es, por su número de resolución, la sentencia de 4 de julio
  (ROJ `STS 3144/2023`) y, por su ROJ, otra de 9 de febrero (resolución 204/2023). Pedida sin más, llegó la
  segunda.
- **«Cita» se mide por el ECLI.** La entrada medía que una respuesta no citara una sentencia inventada por «el
  ECLI, el ROJ o el número de resolución con su fecha», y la respuesta correcta tiene que nombrar ese número
  para decir que no lo ha comprobado. La cita y la línea `⚠ SENTENCIA NO COMPROBADA:` tienen forma fija, y el
  umbral cuenta los ECLI que no vienen de una entrega ni de la persona: un hecho de la sesión (ADR 0037).
- **«Resume» lo decide un juez medido, en un hito propio.** La entrada daba por hecho que `jurisprudencia`
  declaraba la clase `afirma_lo_no_leido` y que decidía. El ADR 0037 exige para eso una rúbrica, casos
  etiquetados y una medida; la rúbrica validada pregunta por preceptos y por el BOE, y no hay casos de una skill
  que no existe. H23 decide con lo que tiene forma, y H25, con la rúbrica validada fuera de un run sobre las
  respuestas del cierre de H23, hace decidir al juez. Ninguna release lleva `jurisprudencia` antes.
- **El texto de una sentencia lo trae una persona, también en las evals.** La eval de «la persona pega el texto»
  necesita un texto real, y un run ni lo descarga ni lo escribe de memoria. Es un fragmento —el encabezamiento y
  el fallo de una sentencia— del documento que una persona descargó del buscador con su navegador:
  `evidencias/adr-0036/`, con la huella del documento, que no se versiona. Con las grabaciones, es lo único del
  CENDOJ que hay en el repositorio, y se mantiene en lo mínimo.
- **Si el CENDOJ bloquea una grabación, no se insiste.** El test de grabación falla con una respuesta que no es
  ni resultados ni «no se ha encontrado»: un CAPTCHA o un 403 no quedan como fixture, el run se detiene y decide
  una persona. La página del CAPTCHA no se provoca para tenerla: en los tests es una página que no se reconoce.
