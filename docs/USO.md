# Bitácora de uso

Qué se le pidió al kit, qué falló y qué faltó. La rellena quien lo usa, en el momento, sin pulir. Es la
entrada del ritual de repriorización (`docs/ROADMAP.md` §6): cada tres o cuatro hitos se relee y el
siguiente hito sale de aquí o del backlog; ante discrepancia gana esta bitácora y el roadmap se actualiza
(ADR 0013).

Una entrada por sesión de uso, la más reciente arriba. Formato libre; basta con que quede claro qué se
quiso hacer y qué pasó. Las referencias a municipios concretos son bienvenidas aquí (es uso, no producto).

## Entradas

### 2026-10-04 · Cuatro pruebas antes de mejorar la instalación: una pieza basta y la versión nueva no llega sola

- **Qué se pidió.** Cuatro pruebas antes de decidir cómo se mejora la instalación, a raíz de un abogado con Windows
  y sin perfil técnico al que Scoop le dio miedo y cuyo equipo bloqueaba cada paso «por contener código peligroso»:
  si la extensión se instala en la app de Claude en Windows; si un plugin con el servidor dentro trae skills y
  herramientas con una sola instalación; si el marketplace entrega solo la versión siguiente del plugin, y con ella
  el servidor; y si firmar el `.mcpb` cambia el aviso rojo. Los días 3 y 4, con la app de escritorio de Claude
  2.19675.0 en macOS, Claude Code 2.1.284 y un plugin de prueba, `kitlegal-prueba`, en un catálogo aparte. La
  pregunta, siempre en una conversación nueva: «¿Qué dice el art. 21 de la Ley 39/2015?».
- **Qué pasó.**
  - **Una sola pieza, con el `.mcpb` dentro del plugin** (`"mcpServers": "./servers/kitlegal.mcpb"`, 41 MB en el
    repositorio del catálogo): la app sincroniza el catálogo y la ficha del plugin enseña «Conectores · 1 ·
    kitlegal.mcpb · Se ejecuta en cada sesión». En una conversación de chat, sin carpeta y con ella, carga
    `boe-legislacion`, pide permiso para las herramientas, llama a `boe_articulo` y `graph_check` del servidor del
    plugin y responde con la cita. El servidor corre en el equipo, lanzado por la app. A Claude Code el plugin le
    llega sincronizado desde la cuenta, sin instalarlo: `/mcp` enseña `plugin:kitlegal-prueba:kitlegal` y la respuesta
    lleva `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]`, leída con la herramienta del plugin.
  - **El servidor no arranca al instalar el plugin.** La app lo conecta en su ciclo de la hora o al reiniciarla (siete
    segundos después de abrir). Seis minutos después de instalar, la conversación tenía la skill y ninguna
    herramienta, y respondió «⚠ SIN CONSULTA AL BOE», sin citar nada de memoria.
  - **Desde el móvil y desde claude.ai en un navegador**, con la app de escritorio abierta en el equipo: la
    conversación usa las herramientas, que corren en el equipo (la caché se escribió allí). Lo mismo con la extensión
    instalada a mano en lugar del plugin. Con la app de escritorio cerrada, el móvil responde «⚠ SIN CONSULTA AL BOE»
    y explica que el ordenador puede estar dormido o con la app cerrada.
  - **El binario de Windows, sin la app**: en un Windows Server 2025 de GitHub Actions, el `kitlegal.exe` del
    `kitlegal.mcpb` de la v0.4.1 da su versión, lee el artículo con el mismo `hash` que en macOS y atiende como
    servidor con sus diez herramientas; Defender no encuentra nada. Era la primera vez que se ejecutaba en Windows:
    todo lo de `ci.yml` y `release.yml` corre en Linux.
- **Qué falló.**
  - **El servidor por URL.** Con `mcpServers` apuntando por `https://` al `kitlegal.mcpb` de una release, la app da
    «Error al sincronizar el marketplace»; el mismo plugin sin esa línea sincroniza. `claude plugin validate` lo
    acepta y la documentación de Claude Code lo describe.
  - **La actualización sola.** Con «Sincronizar automáticamente» activado y la versión siguiente del plugin en el
    repositorio, la app no la trajo en casi catorce horas, abierta toda la noche: sincronizó cada veinte minutos,
    hizo un pase completo cada hora y se reinició una vez, siempre con «0 to download». Llegó con dos gestos: en la
    ficha del plugin, ⋮ > «Buscar actualizaciones», que enciende «Actualizar» en menos de un minuto, y «Actualizar».
    Nada avisa de que hay una versión nueva. Una vez actualizado el plugin, el servidor sí pasó solo de la v0.4.0 a
    la v0.4.1, sin reiniciar, a los 43 minutos; Claude Code la recibió al abrir una sesión y la usó en la siguiente.
  - **La firma.** Un `.mcpb` firmado con `mcpb sign --self-signed` (`mcpb` 2.1.2, la última) no se puede instalar:
    «Error al previsualizar la extensión (…) Invalid comment length. Expected: 2264. Found: 0», que son los bytes del
    bloque de firma detrás del final del zip. Y `mcpb verify` lo da por no firmado, como al original: la
    verificación llama a una función de `node-forge` que no está escrita, con cualquier certificado, y la app lleva
    la misma función.
- **Qué faltó.**
  - **La prueba en Windows con la app.** La persona no pudo hacerla. Siguen sin probar el doble clic, SmartScreen
    sobre el fichero descargado y el antivirus o la política de su equipo, que es lo que le bloqueó. `kitlegal.exe`
    no lleva firma de Windows.
  - **Saber qué versión del servidor corre**: la app no lo enseña. Se vio desde fuera, buscando el proceso y
    pidiéndole `--version`.
  - **La forma de la cita** salió exacta en Claude Code y con variaciones en dos respuestas del chat (todo dentro del
    corchete en una; el rango de la norma en medio en otra). Una respuesta de cada una no mide nada.
  - **Lo que sigue sin probar**: el plugin con el `.mcpb` dentro en Windows, y cuánto pesa en el repositorio del
    catálogo un paquete de 41 MB por release.
- **Qué se corrige de lo anotado antes.**
  - La documentación de Claude dice que el chat ignora el servidor local de un plugin, y la entrada del 2026-10-01
    anotó que una conversación sin carpeta recibía la skill y no las herramientas: en la app de escritorio las
    recibe, una vez que la app ha conectado el servidor. Aquella prueba pudo dar con el mismo retraso.
  - La página `/instalar/` dice que en Claude en la web y en el móvil no funciona, y que el plugin del marketplace
    se actualiza solo: con el equipo encendido y la app de escritorio abierta, la web y el móvil funcionan, y el
    plugin no se actualiza solo.
- **Qué se hizo.** Nada en el producto: estas pruebas deciden qué se construye después. El equipo queda con la
  extensión de la v0.4.1 y el plugin del catálogo real.

### 2026-10-03 · Con la v0.4.1, el marketplace sincroniza en la app de Claude

- **Qué se pidió.** Repetir con la v0.4.1, la que publica el catálogo con el plugin dentro, el paso que falló en la
  aceptación de H22: añadir el marketplace `jmorenobl/kitlegal-plugins` en *Customize > Plugins > Add* de la app de
  escritorio de Claude en macOS, en el modo de chat, instalar desde él el plugin `kitlegal` y preguntar por el art. 21
  de la Ley 39/2015 en una conversación nueva.
- **Qué pasó.** El marketplace sincronizó, el plugin se instaló desde él y la respuesta llevó
  `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]`.
- **Qué falta.** Sigue sin probar que la app entregue sola la versión siguiente del plugin cuando se publique.
- **Qué se hizo.** La web pasa a ofrecer la instalación sin terminal como la oficial, con los botones de descarga de
  las dos piezas y el marketplace como la forma recomendada para el plugin.

### 2026-10-02 · La aceptación de H22 con la v0.4.0: las dos piezas funcionan, y la app no admite el marketplace

- **Qué se pidió.** La aceptación humana de H22 con la v0.4.0, la primera release con las dos piezas: `kitlegal.mcpb`
  y `kitlegal-plugin.zip` descargados de la release con un navegador e instalados en la app de escritorio de Claude en
  macOS, sin abrir una terminal, y «¿qué dice el art. 21 de la Ley 39/2015?» en una conversación nueva sin carpeta. Fue
  en el mismo Mac que desarrolla kitlegal, no en otro: no hay otro. La extensión y el plugin de la prueba de H21 se
  habían desinstalado antes.
- **Qué pasó.** La respuesta transcribió el artículo entero y llevó `art. 21 de la Ley 39/2015 [BOE-A-2015-10565,
  bloque a21]`, el rango de la norma, el recordatorio de que el texto consolidado es informativo y el enlace del
  índice, sin vocabulario interno. macOS no puso ningún reparo al `.mcpb` descargado con el navegador. La ficha de la
  extensión salió con su icono y la descripción entera.
- **Qué falló.** Añadir el marketplace `jmorenobl/kitlegal-plugins` en *Customize > Plugins*: «Error al sincronizar el
  marketplace. Verifica la URL del repositorio e intenta de nuevo». El repositorio es público y su `marketplace.json`
  pasa `claude plugin validate`; su entrada era de fuente `archive`, el zip de la release con su huella. Un
  repositorio de prueba con el mismo plugin dentro, como carpeta, y fuente `./plugins/kitlegal` sí sincronizó.
- **Qué faltó.**
  - **Enterarse de las actualizaciones**: subiendo el zip, quien instala no recibe la versión siguiente ni sabe que
    existe. Por eso el marketplace tiene que funcionar en la app.
  - **El aviso rojo** sale también al añadir un marketplace («los plugins instalados desde tiendas no están
    controlados por Anthropic»), además de en la ficha de la extensión. No depende de nada del manifiesto: solo deja
    de salir con lo que está en el directorio de Anthropic.
  - **Lo que sigue sin probar**: que el marketplace entregue de verdad la versión siguiente, cómo se actualiza la
    extensión, Windows, y Claude en la web con el plugin y sin la extensión.
- **Qué se hizo.** El catálogo pasa a llevar el plugin dentro, publicado por la release, y el README recomienda el
  marketplace para el plugin (ADR 0035, «Prueba con la v0.4.0»).

### 2026-10-02 · El benchmark de observatorio.legal: lo que más pesa es la jurisprudencia, y un ECLI del Supremo sí se puede comprobar

- **Qué se pidió.** Si pedir a observatorio.legal que evalúe kitlegal en su benchmark de asistentes legales, y si es
  el momento.
- **Qué se vio.** La campaña publicada (C2026-Q3, del 7 de abril al 7 de agosto de 2026) mide 15 asistentes con 90
  preguntas por tres réplicas, por la interfaz de cada producto: 30 de fondo, 15 de jurisprudencia, 10 de vigencia,
  10 con referencias falsas, 8 de plazos, 6 de redacción, 6 que no deben rechazarse y 5 ambiguas. El índice da 23
  puntos de 100 a que las citas existan y 9 a que estén verificadas. Claude Fable 5, sin nada más, está en 88,5; el
  primero, en 90,6. Entra quien responde consultas en castellano sobre derecho español y se puede probar «por la vía
  ordinaria»; no publican cómo se solicita ni cada cuánto miden.
- **Qué faltó.**
  - **Jurisprudencia**: kitlegal no consulta ninguna, y es la categoría que alimenta los dos componentes de citas.
  - **Plazos y redacción**: son H9 y H11.
  - **La vía ordinaria**: el día de la consulta, instalarlo pedía una terminal; H22 salió esa misma noche, en la
    v0.4.0.
  - Hoy kitlegal añadiría a Claude sobre todo la vigencia y las citas de legislación, y la diferencia tendría que
    pasar de cinco puntos para contar.
- **Qué se probó.** Que el formulario del CENDOJ resuelve un ECLI con HTTP simple y el agente identificable de
  kitlegal: dos peticiones a mano devolvieron `STS 3144/2023` con su sala, fecha, números y URL, sin navegador y sin
  CAPTCHA. El guion de la skill fiscal anterior usaba un navegador disfrazado de Chrome y buscaba por texto libre; no
  hace falta ni se porta. Se leyó además el marco: el reglamento del CGPJ sobre reutilización está anulado desde 2011
  y la Ley 37/2007 se aplica a las sentencias. Todo en `docs/JURISPRUDENCIA.md`.
- **Qué se descartó.** Un modo de jurisprudencia desactivado por defecto y activado para la evaluación: mide un
  producto que nadie tiene. Y buscar por materia o descargar sentencias de forma automática.
- **Qué se hizo.** ADR 0036, en propuesta: del CENDOJ, resolver una resolución identificada por su formulario; la
  búsqueda por materia la hace la persona con la consulta que le prepara la skill. El hito de jurisprudencia del
  backlog queda descrito con esa pieza. No se pide la evaluación todavía: antes, mejor H20, H8 y H9, que se suman a
  H22, ya publicado. Sí se puede escribir ya al observatorio para preguntar si un kit de skills sobre Claude encaja y cuándo es la
  próxima campaña. Al día siguiente, otra prueba a mano: el ROJ y el número de resolución con su
  fecha dan la misma sentencia, y un ECLI inventado, cero resultados. Pendiente de una persona: aceptar el ADR.

### 2026-10-02 · La prueba a mano de H21 en la app de Claude: con la extensión sola no sale la cita

- **Qué se pidió.** La aceptación humana de H21 en la app de escritorio de Claude: «¿qué dice el art. 21 de la Ley
  39/2015?» en una conversación sin carpeta, con el binario de `main` (`097af64`) dentro de un `.mcpb` hecho a mano e
  instalado con doble clic. La mitad de la app de ChatGPT queda para otro día.
- **Qué pasó.**
  - **Con la extensión sola**: la conversación recibió las herramientas y llamó a `boe_articulo`. El texto fue el
    literal del BOE y dijo que no había avisos de vigencia. Pero no llevó la cita con su forma, añadió un enlace a
    `boe.es` que no venía de ninguna herramienta, habló de «respuesta en caché» y dio por «redacción original» lo que
    dedujo de las fechas del sobre. Las `instructions` del servidor piden la forma de la cita con ese mismo ejemplo: o
    la app no se las pasa al modelo o el modelo no las sigue.
  - **Con la extensión y un plugin solo con las dos skills**, subido como zip en el modo de chat: `boe-legislacion` se
    activó con la pregunta, sin `/`; leyó `normas.md`; llamó a `boe_indice`, `boe_articulo` y `graph_check`; y la
    respuesta llevó `art. 21 de la Ley 39/2015 [BOE-A-2015-10565, bloque a21]`, el recordatorio de que el texto
    consolidado es informativo y ninguna mención de la caché. El enlace a `boe.es` salió de la `url` del índice.
    Resumió los seis apartados con títulos suyos en lugar de transcribirlos.
- **Qué faltó.**
  - **La ficha de la extensión**: sin icono, con la descripción cortada en la cabecera, y con el aviso rojo de la app
    —«otorgará a esta extensión acceso a todo lo que hay en tu computadora», desarrollador sin verificar por
    Anthropic—, que a quien no es técnico le va a parar.
  - **Lo que sigue sin probar**: los dos ficheros descargados con un navegador en otro Mac (macOS no puso reparos, pero
    el `.mcpb` se había creado en el mismo equipo), cómo se actualizan, Windows, y la web y el móvil.
- **Qué se hizo.** H22 detallado en el roadmap con dos piezas por release —la extensión y el plugin—, la ficha con
  icono y descripción corta, y el README explicando el aviso. Lo que sigue sin probar es la aceptación humana de H22.

### 2026-10-01 · Usar kitlegal desde Claude Cowork: hoy no se puede

- **Qué se pidió.** El 2026-09-28: cómo usar kitlegal desde Claude Cowork, y si hacía falta un servidor MCP.
  Después, si poner un servidor propio para llegar a ChatGPT y si habría que cobrarlo.
- **Qué falló.** Nada del producto: en Cowork no hay forma de usarlo. Cowork carga las skills y los plugins de la
  cuenta, no los del disco, así que `kitlegal skills install` no le llega; y aunque la skill llegara, dice «ejecuta
  `kitlegal boe articulo …`» y el shell de Cowork es una máquina virtual donde el binario del equipo no está.
- **Qué faltó.**
  - **Un servidor MCP local.** Es lo que cruza esa frontera: el programa corre en el equipo y el agente recibe sus
    operaciones como herramientas. La documentación de Claude dice que el servidor local de un plugin carga en
    Cowork cuando la sesión corre en el equipo.
  - **Skills que sepan usar herramientas.** Las dos están escritas para un shell (`… && kitlegal graph check …`,
    códigos de salida) y las evals solo miden órdenes.
  - **Una instalación sin terminal.** Un plugin puede llevar las skills y un paquete `.mcpb` con el binario dentro,
    y se instala desde *Customize > Plugins*. Probado a mano el mismo día con un plugin de prueba, en la app de
    Claude para macOS: subido como zip en el modo de chat —que ya es el mismo que Cowork—, la app arranca el
    binario sin notarizar en el equipo, el servidor lee el BOE y escribe en `~/.cache`, y la skill lo usa en una
    conversación que tiene elegida una carpeta del equipo. Sin carpeta, la conversación recibe la skill y no las
    herramientas; con el mismo `.mcpb` instalado suelto, con doble clic, sí las recibe. Subido estando en el modo Code, el chat no lo ve. Al instalarlo, la app avisa en rojo de que «otorgará acceso a todo
    en tu computadora». Y «prueba kitlegal» no activó la skill: Claude entendió que había que pasar los tests del
    proyecto.
  - **ChatGPT.** En la web y en el móvil solo admite servidores remotos. La app de escritorio sí admite servidores
    locales por stdio, con el binario instalado.
- **Qué se hizo.** Decisión de Jorge: todo en el equipo, con todas las funciones, y el servidor MCP por delante de
  H20. El servidor remoto queda aplazado: tiene que pensarlo, porque kitlegal pasaría a recibir las preguntas de
  quien lo usa (ADR 0027) y pagaría un servidor que otros consumen sin coste; donde un servidor local no llega se
  dice que no es compatible. ADR 0035: H21 (`kitlegal mcp serve`, las skills con orden y herramienta, las evals en
  los dos modos) y H22 (la extensión de escritorio con el servidor y el plugin con las skills), que se detalla al cerrar H21
  (detallado el 2026-10-02, entrada de arriba).

### 2026-10-01 · La web no respondía a nada que la gente busque, y ninguna de sus dos audiencias usa una terminal

- **Qué se pidió.** Analizar cómo busca el usuario tipo de kitlegal para enfocar el mensaje de la web, sin tocar el
  repositorio; y después, aplicar lo que saliera antes de seguir con el siguiente hito.
- **Qué falló.** Search Console no daba ni una impresión en 28 días con las cuatro páginas indexadas, así que no
  había datos propios. Con el autocompletado de Google en España (unas 200 búsquedas de arranque) y los estudios
  publicados entre 2024 y 2026 salió que la web contaba lo que hace el producto y la gente busca lo que le pasa: el
  plazo («cuánto tiempo tengo para…»), el problema en primera persona («mi casero no me devuelve la fianza», «el
  ayuntamiento no contesta a un escrito») y el siguiente paso («modelo recurso de alzada»). Los despachos, que son de
  una a tres personas (Libro Blanco del CGAE, 2026), buscan «ia para abogados gratis», «… españa» y «claude para
  abogados»; las multas de los tribunales han sido por jurisprudencia inventada, que kitlegal no comprueba, y no por
  normas derogadas, que era el titular de la portada. Y la sección «Sin barreras técnicas» no era verdad para
  ninguna de las dos audiencias.
- **Qué faltó.**
  - **Una instalación sin terminal.** Menos del 2 % de la población usa Claude con frecuencia (Funcas, enero de
    2026) y los abogados buscan «claude cowork abogados», no Claude Code. Un servidor MCP local no lo arregla por sí
    solo, porque sigue haciendo falta el binario: lo que quitaría la terminal es un paquete de un clic con el binario
    dentro, sin comprobar todavía (firma en macOS). Es la mayor distancia entre la web y quien la lee, y no se
    arregla con texto.
  - **Comprobar cada noche que las citas de la web siguen vigentes.** Al regenerar los sobres, el artículo 36 de
    la Ley de Arrendamientos Urbanos tenía una versión del 30 de septiembre de 2026 (Real Decreto-ley 26/2026) con
    un apartado 7 nuevo, y la web publicada seguía con la de 2019. Los sobres solo se regeneran cuando alguien
    ejecuta `make web-citas`.
  - **El último día de un plazo.** Las consultas más buscadas piden una fecha («hasta qué día tengo»), y hoy la
    web solo puede decir cómo se cuenta: es H9 (`plazos`).
  - **El escrito.** «Modelo recurso de alzada» y «modelo recurso de reposición» están entre las sugerencias más
    repetidas: es H11.
  - **Las multas de tráfico.** Recurrir una multa (ZBE, SER, DGT) es lo más buscado como acción y queda fuera de la
    base por la disposición adicional primera de la Ley 39/2015 (ADR 0027): la web no lo ofrece. Si el uso lo pide,
    es la vertical de tráfico.
- **Qué se hizo.** Decisión de Jorge, en sesión interactiva y fuera del workflow `hito` (ADR 0034): la web habla al
  usuario final, ciudadanía y despachos, no al perfil técnico ni a las administraciones públicas; la portada del
  sitio es la de la ciudadanía y despachos pasa a `/despachos/`, con una invitación a escribir a `info@kitlegal.es`
  y lo que hoy no hace dicho ahí; seis páginas en `/consultas/` con la pregunta como se busca y cada afirmación con
  su cita al lado; cada cita muestra los avisos de su sobre y declara la versión del bloque con la que se escribió,
  y la construcción falla si el BOE trae otra; `/instalar/` pasa a ser una guía paso a paso; y `docs/SEO.md` deja
  los directorios para desarrolladores.

### 2026-09-30 · El alias `sonnet` ya es Sonnet 5.5 y el job de evals sigue midiendo Sonnet 5

- **Qué se vio.** Claude Code 2.1.284 resuelve el alias `sonnet` a `claude-sonnet-5-5` (2.1.283 y anteriores, a
  `claude-sonnet-5`), y la resolución es del cliente: la misma orden con 2.1.270, la versión del job, sigue dando
  Sonnet 5, y 2.1.270 avisa `[claude-code:unrecognized_model]` si se le pide el id nuevo. El job decidía con
  `claude-sonnet-5`, que ya no es el modelo del uso real (ADR 0016). Además, su control del modelo de la sesión
  aceptaba cualquier id que empezara por el pedido, así que no distinguía `claude-sonnet-5` de `claude-sonnet-5-5`.
- **Qué se midió.** Sondeo comparado sobre `941b24f` con 2.1.284, las 19 evals de `boe-legislacion` y las 3 de
  `legal-core`, tres repeticiones (ADR 0031, «Medidas»):
  - **Sonnet 5**: todas las series que deciden 3 de 3; 3 de 51 respuestas con expresión prohibida; 296 s.
  - **Sonnet 5.5**: la 04 (LGT, art. 66), que decide, 1 de 3, porque dos sesiones no activan la skill y responden de
    memoria; 0 de 51 con expresión prohibida, pero 6 de 51 primeras líneas dicen el estado de lo comprobado con
    palabras que la lista no tiene («no trae avisos de vigencia», «Lo he leído en el BOE consolidado»); 217 s.
- **Qué se hizo.** ADR 0031: decide `claude-sonnet-5-5` con Claude Code 2.1.284, fijados juntos; Sonnet 5 sale del
  job; el control del modelo exige el id pedido o ese id con su fecha; y antes de cada hito una orden comprueba si el
  alias ha cambiado. La activación de la 04 y el anuncio de lo comprobado son de la skill con el modelo nuevo: los
  arregla H7.4.

### 2026-09-30 · H7.3 cumple su umbral con la lista, pero no con la lectura: la comprobación contada con otras palabras y una redacción que nadie leyó

- **Qué se pidió.** Leer, respuesta a respuesta, lo que dijo `boe-legislacion` en las tres mediciones del cierre de
  H7.3 (#85): sobre `6ab3add` (5 de 51 con la lista, 571 s), `eb6b4c8` (0 de 51, 540 s) y `196ee05` (1 de 51, 492 s;
  `specs/013-h7-3-el-umbral-de/gates/evals/boe-legislacion.json`, el único versionado en `main`; los otros dos
  siguen en los commits de la rama, fuera de `main`). Todas las cifras son de las respuestas del modelo que decide en
  las evals que activan la skill; el modelo informativo no tiene ninguna respuesta de las dos clases de abajo en
  ninguna medición. Las sesiones se nombran `<eval>-<sesión>`.
- **Qué falló.**
  - **La comprobación, contada con otras palabras (clase A).** Respuestas que dicen el estado de la comprobación de
    la redacción o de una lectura anterior —que cambió o que no, que no hay nada que señalar, que la comprobación
    terminó— o anuncian lo que el agente tiene o va a hacer, al principio o al final de la respuesta y en cualquier
    idioma. En `196ee05`, 7, de las que la lista marca una (14-03): 01-02 «No hay avisos de vigencia sobre este
    bloque ni cambios de redacción respecto a una consulta anterior.»; 05-02 «…y no consta que su redacción haya
    cambiado desde una consulta anterior.»; 13-02 «…ni indicios de redacción posterior a la consultada.»; 13-03 «No
    se ha detectado ningún cambio de redacción respecto a una lectura anterior de este bloque.»; 14-01 «Article 59
    already answers the question fully, so I don't need art. 60. I have enough to respond now.»; 14-02 «Con el
    artículo 59 tengo suficiente para responder a la pregunta completa; no necesito el artículo 60 para esto.»; y
    14-03 «La comprobación de redacción terminó sin hallazgos, así que no hay cambios que señalar.». En `6ab3add`,
    11 (la lista, 5; además 01-03, 03-01, 03-03 «Sin cambios de redacción respecto a lecturas anteriores. Aquí está
    la respuesta.», 13-03 «Sin cambios desde una lectura anterior. Respondo.», 14-03 y 15-03); en `eb6b4c8`, 4 (la
    lista, 0: 06-01, 14-01 «Este artículo ya da la respuesta completa. Sin avisos de vigencia ni cambios desde una
    lectura anterior.», 15-02 y 17-01). Es el vocabulario de `SKILL.md` v0.1.3: «desde una lectura anterior» sale
    seis veces en su prosa (líneas 66, 74, 83, 177, que es el título de una sección, y 270) y «lectura anterior»,
    diez. «No hay avisos de vigencia sobre este bloque», sola, no es de esta clase: es derecho (entrada del
    2026-09-29 sobre H7.1).
  - **Una redacción que ninguna orden devolvió (clase B).** En la eval 19, 5 de las 9 sesiones de las tres
    mediciones describen la redacción superada del art. 118 LCSP —qué decía o qué cambió—: 19-01 de `6ab3add`,
    19-01 y 19-02 de `eb6b4c8`, y 19-01 y 19-02 de `196ee05`. Ninguna orden de esas sesiones la devolvió: `boe articulo BOE-A-2017-12902 a1-30` da solo la redacción vigente, con `fecha_vigencia`
    20200206 y `norma_modificadora` `BOE-A-2020-1651`, y `graph check`, dos fechas (FR-070 de H7). Dos son falsas
    contra la respuesta grabada, que trae las dos redacciones: 19-02 de `196ee05` dice que la anterior estuvo
    «vigente hasta el 9 de marzo de 2018» (20180309 es cuándo entró en vigor) y que «no distinguía umbrales entre
    obras y suministros/servicios del mismo modo» (el apartado 1 original ya fijaba 40.000 y 15.000 euros), y 19-01
    de la misma medición, que «ya no exige tres informes separados»; 19-02 de `eb6b4c8` atribuye a la original un
    informe «motivando la elección del contratista entre varios presupuestos» que no tenía. Es contenido legal sin
    fuente (constitución, principio II), acierte o no. Ya pasaba en los informes versionados de H7.2 (19-01, «Si lo
    que tenías en mente era la redacción original de 2018 (por ejemplo, sin el requisito específico de motivar…»,
    sin marcar; y 19-02, «El cambio relevante fue que se eliminó la exigencia de un informe adicional…», marcada por
    «hallazgo») y de H7.1 (19-01, «el contenido normativo del artículo 21 no ha cambiado por una reforma
    posterior»). `SKILL.md` ya dice «Lo único que la respuesta dice de una lectura anterior es esa línea»: la
    prohibición no basta.
  - **El recuento real.** Contando las dos clases, `196ee05` da 9 de 51 (17,6 %; la lista, 1), `eb6b4c8` 6 de 51 y
    `6ab3add` 12 de 51. Con la etiqueta de arriba, en los informes versionados llevan alguna expresión o alguna de
    las dos clases: en H7.1, 36 de 93 (las 35 de la lista y 19-01); en H7.2, 11 de 93 (las 10 y 19-01); en H7.3
    (`196ee05`), 9 de 93. Buscadas con `jq` las formas de las dos clases en las respuestas sin marcar de los tres,
    no hay más.
  - **El job juzga otra cosa que la respuesta.** En `eb6b4c8`, la sesión 03-01 lanzó `kitlegal boe buscar` en
    segundo plano, y el último `result` del transcript, el que juzga el job (`internal/evals/sesion.go`,
    `leerResult`), fue la réplica al aviso de esa tarea: «Esa tarea en segundo plano era solo una búsqueda
    auxiliar… No afecta a la respuesta ya entregada…». La sesión falló por cita ausente y contó como respuesta
    limpia en el total de 51. Y `recontarExpresiones` (`internal/evals/informe.go:762`), que usan el job y el
    sondeo, cuenta como limpia una sesión que no terminó; en estas mediciones no hubo ninguna.
  - **Dos ejecuciones del job por commit.** Sobre `6ab3add`, la apertura de la propuesta de cambio y la etiqueta
    `evals` lanzaron dos ejecuciones con 4 s de diferencia (36672667544 y 36672671529): una terminó a los 12 min y
    la otra esperó a que terminara y volvió a medir el mismo commit, 35 min después de abrirse.
  - **El sondeo, con la traza de `go test`.** Sin `CLAUDE_CODE_OAUTH_TOKEN`, o con un argumento que no vale,
    `TestSondeo` falla con `require.NoError` y `scripts/evals-sondeo.sh` imprime el registro entero de `go test`:
    el mensaje llega dentro de la traza de testify.
  - **`&&` en Windows.** La documentación de Claude Code dice que en Windows nativo la herramienta PowerShell se
    activa sola sin Git Bash, y que con Git Bash está activada por defecto en las cuentas de claude.ai y es la
    shell principal; busca `pwsh.exe` (PowerShell 7) y, sin él, usa `powershell.exe` (5.1)
    (<https://code.claude.com/docs/en/tools-reference#powershell-tool>, <https://code.claude.com/docs/en/setup>).
    Windows PowerShell 5.1 no tiene `&&`, y la orden de lectura y comprobación de `SKILL.md` lo usa. Sin medir: el
    job corre en Linux.
- **Qué faltó.** Que la skill deje de contar la comprobación y de describir lo que no ha leído, y que la lista mida
  las dos clases y no sus frases; una eval que lea dos bloques de una norma con redacción cambiada; que el job juzgue
  la respuesta a la pregunta y una sola vez por commit; y que el sondeo diga sus errores de uso sin traza.
- **Qué se hizo.** Abrir H7.4, entre H7.3 y H20, antes de la release.

### 2026-09-29 · H7.2 no cumple su umbral: el ruido cambia de palabras, y medirlo cuesta 40 minutos

- **Qué se pidió.** Verificar el cierre de H7.2 (#82) contra su SC-001: como mucho el 5 % de las respuestas de cada
  modelo en las evals de `boe-legislacion` que activan la skill llevan una expresión prohibida. Informe del job sobre
  `fbdab2e` (`specs/012-h7-2-la-consulta-repetida/gates/evals/boe-legislacion.json`, clave
  `expresiones_prohibidas_por_modelo`).
- **Qué falló.**
  - **El umbral no se cumple.** Sonnet 5: 10 de 51 (19,6 %; H7.1, 34 de 51; H7, 43 de 51). Haiku 4.5: 0 de 30. Por
    eval: 03 (1), 06 (1), 13 (2), 14 (2), 15 (3) y 19 (1); ocho son de evals informativas, que no deciden nunca, y las
    de la 03 y la 06 están en series que pasaron con 2 de 3. `legal-core`, 0 de 18.
  - **Una sola forma, con las palabras de `SKILL.md`.** Ocho empiezan por «Sin hallazgos que trasladar.» —a veces con
    «Ya puedo responder.», «Redacto la respuesta.» o «Con esto ya tengo la respuesta completa»—, y otra por «No hay
    hallazgos, así que respondo con el texto vigente.». Es el párrafo de transición tras la última orden
    (`graph check`) que ya describía el research de H7.2, dicho ahora con el vocabulario del protocolo: la regla 7 de
    v0.1.2 se titula «Una comprobación con hallazgos no es un fallo» y dice «con hallazgos o sin ellos […]
    trasládalos», y el paso 5 enumera «los hallazgos» entre lo que no se nombra. Las demás respuestas de Sonnet 5 ya
    no llevan ningún párrafo de transición, ni con otras palabras: el ruido se ha concentrado, no desplazado. Las
    frases de anuncio («ya puedo responder», «ya tengo todo lo necesario», «redacto la respuesta», «que trasladar»)
    solo aparecen, en los cierres de H7.1 y H7.2, en respuestas que la lista ya marca.
  - **Un ejemplo copiado como dato.** La eval 19, sesión 02 de Sonnet 5, empieza por «⚠ REDACCIÓN MODIFICADA: la
    redacción con fecha de vigencia 20180309, la que se consultó antes, ha sido sustituida por la de 20250101... en
    realidad las fechas del hallazgo son 20180309 y 20200206». 20250101 y «la que se consultó antes» salen del ejemplo
    de «Memoria de consultas» de `SKILL.md`, que describe una redacción del art. 21 LPAC que el BOE no tiene.
  - **Todo salió en verde.** El job dio «aprobado» porque la lista decide por serie (2 de 3) y el 5 % solo se
    publicaba; el cierre no intentó arreglarlo y la trazabilidad dio SC-001 por hecho. Lo corrige para el proceso el
    ADR 0029; el job y la skill, H7.3.
  - **Cada intento cuesta 40 minutos.** El trabajo de `boe-legislacion` abre sus 93 sesiones una tras otra: 33 min 22 s
    y 34 min 28 s de sesiones en las dos ejecuciones del cierre (36 min 37 s y 42 min 33 s el trabajo entero), unos
    22 s por sesión; `legal-core`, 18 sesiones en menos de 4 min. Exige Linux con strace y sudo, así que un cambio de
    `SKILL.md` no se puede probar en un Mac. Además, cada cierre ejecuta el job dos veces a la vez sobre el mismo
    commit (el evento de apertura y la etiqueta `evals`; también en H7 y H7.1): el doble de sesiones sin medir nada
    nuevo. Esas dos ejecuciones simultáneas son, a la vez, lo único medido sobre cuántas sesiones aguanta la
    suscripción a la vez: dos de `boe-legislacion` durante media hora y cuatro durante unos minutos, con las 111
    sesiones de cada una terminadas con `result success`.
- **Qué faltó.** Que el umbral decida en el job; que `SKILL.md` no enseñe el vocabulario que la respuesta no puede
  decir ni lleve ejemplos con datos inventados; y un modo de medir un cambio de la skill en minutos, en local, antes de
  pedir el veredicto.
- **Qué se hizo.** Abrir H7.3, entre H7.2 y H20, y no publicar ninguna release hasta que el umbral se cumpla.

### 2026-09-29 · La eval de la consulta repetida es imposible, y las respuestas narran la comprobación

- **Qué se pidió.** Verificar el cierre de H7.1 (#80) leyendo lo que responde `boe-legislacion` en las evals: el informe
  del job sobre `a8aeb0d` (`specs/011-h7-1-graph-check-acotado/gates/evals/`) y, para comparar, el del cierre de H7
  sobre `d4a96db` (ejecución 36431889063, sacado del registro del job con `gh run view … --log | cut -f3-`).
- **Qué falló.**
  - **La eval 19 plantea una situación imposible.** Su grafo previo siembra una redacción «anterior» del art. 21 LPAC
    derivada a mano, con vigencia 20151002, anterior a la entrada en vigor de la ley. La caché sirve la respuesta real
    del BOE, que trae una sola `<version>`: la original, con vigencia 20161002. El grafo y el BOE se contradicen.
    Sonnet 5 pasa 3 de 3 en H7 y 2 de 3 en H7.1. Una sesión de H7.1 que traslada la forma añade que lo consultado
    antes «era la versión previa a su entrada en vigor» y que el contenido no ha cambiado: vio la incoherencia. La que
    falla (`19-…-claude-sonnet-5-03`) no traslada `version-obsoleta`, dice que el artículo «no ha sido modificado por
    ninguna norma posterior» y remata con «Este texto coincide con lo que te habría confirmado antes», algo que la
    skill no puede saber. En la realidad esa contradicción no se da: cuando el BOE consolida un cambio, la respuesta del
    bloque trae sus redacciones anteriores y la norma que lo modificó. Un artículo que lo cumple es el art. 118 de la
    LCSP (`BOE-A-2017-12902`, bloque `a1-30`): el Real Decreto-ley 3/2020 lo modificó, y su respuesta grabada en H4 trae
    las dos redacciones (vigencia 20180309 y 20200206).
  - **Las respuestas empiezan contando la comprobación.** «Sin hallazgos en la memoria de consultas. Ya tengo todo lo
    necesario para responder.», «Código 0 sin hallazgos, así que no hay nada que trasladar de la memoria de
    consultas.», «Sin hallazgos ni avisos de vigencia.»… `SKILL.md` lo prohíbe desde H7 («Con código 0 y sin
    `version-obsoleta`, no digas nada de la memoria de consultas»), y ninguna eval lo mide. Medido con este patrón, sin
    distinguir mayúsculas, sobre la respuesta entera:

    ```text
    hallazgo|memoria de consultas|graph\s+(check|show|stats)|kitlegal\s+graph|version-obsoleta|fuente-caducada|c[oó]digo(\s+de\s+salida)?\s+[0-7]\b|c[oó]digo\s+de\s+salida|\bjson\b|exit\s+code
    ```

    | Cierre | Respuestas con ruido | En el primer párrafo | Sonnet 5, evals que activan | Haiku 4.5 | `legal-core` |
    |---|---|---|---|---|---|
    | H7 (`d4a96db`) | 43 de 93 (46 %) | 41 | 43 de 51 (84 %) | 0 de 30 | 0 de 18 |
    | H7.1 (`a8aeb0d`) | 34 de 93 (37 %) | 34 | 34 de 51 (67 %) | 0 de 30 | 0 de 18 |

    En H7.1, por eval: 02 (1), 03 (3), 04 (3), 05 (3), 06 (3), 07 (3), 08 (2), 09 (2), 13 (3), 14 (3), 15 (3), 16 (2)
    y 17 (3). Los términos: «hallazgo» en las 34, «memoria de consultas» en 24, «código 0» en 7 y `version-obsoleta` en
    2. Además, una respuesta afirma lo dicho en otra conversación: la de «te habría confirmado antes», en la eval 19
    (patrón `\bte\s+(dije|respond[ií]|confirm[eé]|indiqu[eé]|coment[eé]|contest[eé]|expliqu[eé])\b|\bte\s+habr[ií]a\b|como\s+(ya\s+)?te\b|conversaci[oó]n\s+anterior`);
    en H7, ninguna. En total, 35 respuestas de H7.1 incumplen. En H7, una de las 43 lo dice fuera del primer párrafo
    («mi memoria de consultas de esta herramienta registra…», en la eval 19) y otra da el código de salida de un límite
    de ritmo («(código 5)»). No se cuentan «No hay avisos de vigencia sobre este bloque» ni frases parecidas (dos en
    H7.1): dicen a quien lee que la norma no está derogada, que es derecho y no maquinaria.
- **Qué faltó.** Que la eval de la consulta repetida siembre una redacción anterior que el BOE tenga, derivada de su
  respuesta y no escrita a mano, y que un control lo impida; que alguna eval mire la respuesta en busca de lo que la
  skill prohíbe decir; y saber por qué Sonnet 5 narra la última orden pese a la regla.
- **Qué se hizo.** Abrir H7.2, entre H7.1 y H20 y antes de cualquier release que lleve H7.

### 2026-09-28 · `graph check` devuelve todo lo consultado y repite señales que ya se dieron

- **Qué se pidió.** Validar la rúbrica nueva de los jueces (ADR 0028) sobre H7 recién fusionado (#77), midiendo en el
  binario lo que `boe-legislacion` lee de `kitlegal graph check` con un uso sostenido.
- **Qué falló.** Con un `world.db` sembrado por el propio almacén —300 normas, 2 400 bloques, 240 de ellos con dos
  redacciones, el 90 % consultado hace más de una semana—, `kitlegal graph check --json` sale con 0 y 3 045 524 bytes:
  4 812 hallazgos (4 572 `fuente-caducada` y 240 `version-obsoleta`, unos 633 bytes cada uno). La skill lo pide dos
  veces por pregunta: unos 6,1 MB, cuando lo que atañe a la norma de la pregunta son unos 20 hallazgos (≈ 13 KB). Y
  ninguna señal se apaga: `version-obsoleta` se da para siempre sobre cada redacción superada, y `fuente-caducada`
  también sobre redacciones superadas que ninguna lectura renueva. Al día siguiente de leer la redacción nueva, la
  misma pregunta volvería a decir que la redacción ha cambiado. Pasó por cinco jueces sin que ninguno lo viera, porque
  ningún criterio miraba el uso (ADR 0028, problema B).
- **Qué faltó.** Que `graph check` se pueda acotar a la norma y a los bloques de la pregunta, con una salida que no
  crezca con lo acumulado; que cada señal deje de darse cuando ya no dice nada nuevo; y que la eval de la consulta
  repetida compruebe que la respuesta dice que la redacción cambió.
- **Qué se hizo.** Medirlo (juez B de la revisión en la validación del ADR 0028) y abrir H7.1, entre H7 y H20, que
  además retira de H7 lo que la rúbrica nueva dice que no pasa el umbral de materialidad.

### 2026-09-27 · El roadmap describe un kit municipal; la web y el README ofrecen uno para cualquier asunto

- **Qué se pidió.** Revisar la fase 2 del roadmap, «Actuar en mi municipio»: kitlegal ha acabado siendo una
  herramienta genérica, la web habla a despachos y a la ciudadanía, y un abogado o un ciudadano tendría que poder
  usarlo en cualquier materia —fiscal, laboral, mercantil— de forma genérica, dejando la especialización para las
  verticales. Y contrastarlo con una conversación de Jorge con otro agente sobre lo que existe en el mercado y una
  estrategia para ponerse por delante.
- **Qué falló.** Nada del producto. El roadmap era lo único que seguía diciendo municipio: H10 guardaba «el municipio
  de la persona usuaria», que no sirve a un despacho y no es lo que mira el art. 30.6 LPAC (el del interesado y el de
  la sede del órgano); H11 recurría «un acto municipal» y sacaba el órgano de `territorio`, que solo da el DIR3 de los
  ayuntamientos. Y ese texto es la entrada del run, así que H10 y H11 habrían salido municipales. Al generalizar faltaba
  la frontera con las verticales: la DA 1ª.2 LPAC saca del procedimiento común los tributos, la Seguridad Social, las
  sanciones tributarias, sociales y de tráfico y la extranjería, y un recurso de la LPAC contra una multa de tráfico
  es un escrito mal fundado. De la estrategia externa, dos piezas encajaban y no estaban: la redacción de un artículo
  a una fecha (el BOE ya la da y el binario se queda con la última) y revisar las citas de un texto ajeno. Otras
  chocaban: la telemetría, con la promesa de la web; CENDOJ, con el ADR 0003. Y la conversación daba por hecho que
  faltaban Homebrew y `.deb`, que ya existen.
- **Qué faltó.** Una medida antes de subir de prioridad la búsqueda del artículo por su materia (entrada del
  2026-09-15): con Sonnet, las cinco preguntas por materia de H5 pasaron 3 de 3 (ADR 0016), así que no está demostrado
  que la herramienta haga falta.
- **Qué se hizo.** Decisión de Jorge, en sesión interactiva y fuera del workflow `hito` (ADR 0027): la fase 2 pasa a
  ser «actuar: llevar un asunto ante cualquier administración» y cubre el procedimiento común y el acceso a la
  información en cualquier materia, declarando no cubierto lo que la ley regula aparte; H10 guarda los municipios del
  asunto; H11 genera acceso, alzada o reposición y alegaciones ante cualquier órgano; entra H20, la redacción a una
  fecha, entre H7 y H8; H8 valida textos ajenos; la fase 3 pasa a ser «consultar lo que hacen las administraciones,
  empezando por el municipio»; el artículo por materia queda en el backlog con condición de entrada medible; sin
  telemetría; los directorios, el benchmark y los acuerdos con colegios o universidades quedan fuera del roadmap.
  Constitución 2.5.0.

### 2026-09-27 · La salida de `skills install` es críptica para quien no es técnico

- **Qué se pidió.** Instalar las skills en la cuenta con `kitlegal skills install -g`, siguiendo el README, y saber
  qué había pasado: dónde han quedado, qué agente las verá y qué hacer a continuación.
- **Qué falló.** La orden funciona, pero sin `--json` imprime la tabla mínima genérica del sobre: `fuente
  kitlegal.skills`, `url kitlegal:applet/skills`, `hash sha256:…` y después `0.enlaces.0.host claude`,
  `0.enlaces.0.modo enlace`, `0.estado actualizada`… con la ruta absoluta de HOME en cada línea. A quien no sabe qué
  es un sobre ni una huella no le dice nada, y `skills list` y `skills doctor` salen igual: un `doctor` sin hallazgos
  no dice «todo en orden», y uno con hallazgos da `hallazgos.0.clase`, `hallazgos.0.orden`, en vez de una frase y la
  orden lista para copiar.
- **Qué faltó.** Un camino para que un applet cuente su resultado a una persona sin que el presentador conozca applets
  (ADR 0005) y sin cambiar ni un byte de `--json`, que es lo que leen los agentes y las evals (ADR 0023).
- **Qué se hizo.** Decisión de Jorge, en sesión interactiva y fuera del workflow `hito`, como los ADR 0023 y 0025:
  `schema.Resultado` gana `Legible`, que el kernel escribe sin `--json` en lugar de la tabla (ADR 0026), y `skills`
  lo rellena en `install`, `list` y `doctor`: dónde están las skills y quién las lee de ahí, cada una con su estado o
  su versión y sus entradas por marca (Claude Code, Antigravity), `~` en las rutas, el agente que no las verá con la
  orden que lo enlaza, «todo en orden» o cada hallazgo en una frase con su orden debajo. `boe` y `territorio` siguen
  con la tabla: ahí la procedencia es la cita. `--dry-run` usa el mismo vocabulario.

### 2026-09-27 · Un botón de «instalar» para quien no usa la terminal

- **Qué se pidió.** Si se puede instalar kitlegal con un clic en las apps de escritorio —Claude, Codex, Antigravity,
  Gemini—, sin `curl` ni `kitlegal skills install`, para alguien no técnico.
- **Qué falló.** Al revisar qué lee cada agente salieron tres cosas de lo ya publicado: `kitlegal skills install -g`
  no llegaba a Antigravity, que en global lee `~/.gemini/config/skills/` y no `~/.agents/skills/` (comprobado con
  `agy` 1.2.10); el README prometía Claude Cowork, que carga las skills de la cuenta y no las del disco; y Codex
  ejecuta `kitlegal` sin red, así que pide permiso en cada consulta al BOE.
- **Qué faltó.** No hay botón universal. Todos los «un clic» que existen instalan servidores MCP (`.mcpb` de Claude
  Desktop, enlaces de VS Code y Cursor) o plugins desde un marketplace (Claude: *Customize > Plugins*, que se sincroniza
  con Cowork y Claude Code; Codex: `codex://plugins/install/…` con el marketplace ya añadido), y para llevar el
  binario hay que empaquetarlo por plataforma (falta comprobar si macOS exige firmarlo y notarizarlo dentro de un
  `.mcpb`). La vía es `kitlegal mcp serve` más un plugin, del backlog de distribución. La app Gemini no
  ofrece skills en el Espacio Económico Europeo, y Gemini CLI se retiró para particulares en favor de Antigravity.
- **Qué se hizo.** Decisión de Jorge: por ahora, `curl … | sh` y `kitlegal skills install`. Se arregla lo que falló:
  host `antigravity` para `-g` (ADR 0025), el README deja de prometer Cowork y explica la regla de Codex.

### 2026-09-20 · Instalarlo en otro proyecto, sin clonar, y actualizarlo con cada versión

- **Qué se pidió.** Cómo instalar kitlegal para usarlo en cualquier proyecto —en Claude Code, Codex o Antigravity— y
  actualizarlo conforme salgan versiones, con la menor fricción posible para alguien distinto de quien desarrolla.
- **Qué falló.** Nada del producto: `make install` hace lo que dice. Pero es una instalación para quien desarrolla:
  exige clonar y tener Go, la skill instalada es el árbol de trabajo del clon (cambia con la rama en la que esté) y
  `scripts/boe` es un enlace a un binario de fuera de la skill, que ningún zip ni ningún gestor de paquetes puede
  llevar. No había release (H19 estaba al cierre de la fase 3) y el repositorio es privado.
- **Qué faltó.** Un artefacto instalable sin clonar. Se descartó repartir bundles `.skill`/zip: el binario pesa 11 MB
  comprimido por plataforma, en macOS un ejecutable extraído de un zip descargado queda en cuarentena, y un bundle
  fino con un shim que descargue el binario mete un descargador propio y dos versiones (skill y binario) que pueden
  divergir.
- **Qué se hizo.** Decisión de Jorge (2026-09-20, ADR 0019): el binario se instala con el gestor de paquetes de cada
  plataforma (Homebrew en macOS y Linux, Scoop en Windows, `.deb`/`.rpm`, `install.sh`, `go install`), lleva las
  skills dentro y las instala él con `kitlegal skills install`: local por defecto en `./.agents/skills/`, `-g` en
  `~/.agents/skills/`, y los hosts como enlaces relativos (`--host claude` → `.claude/skills/`), detectados por su
  directorio de configuración cuando no se indica. Es el disparador del ADR 0013: H19 se adelanta a la fase 1, detrás
  de H6, con ese contrato. Entregado en H19 (#47) y publicado como v0.1.1 el 2026-09-27: en un directorio vacío,
  `curl … | sh` y `kitlegal skills install` bastan para que Claude Code responda al artículo 21 de la Ley 39/2015 con
  su cita.

### 2026-09-16 · Ninguna eval comprueba qué hace la skill ante una norma derogada

- **Qué se pidió.** Nada en concreto: la duda salió al explicar por qué el job de evals corre sin red. Si las
  respuestas del BOE están grabadas y la caché congelada, ¿cómo sabe uno que la ley que cita sigue viva?
- **Qué falló.** Nada. El binario hace lo que hay que hacer: `internal/source/boe/avisos.go` deriva de los metadatos
  los avisos `derogada` («⚠ NORMA DEROGADA: esta norma ha sido derogada.»), `vigencia-agotada` y
  `consolidacion-no-finalizada`, y `articulo.go` no emite ningún artículo con la vigencia sin comprobar: si los
  metadatos fallan, la invocación falla con «el bloque X se obtuvo, pero no se pudo comprobar su vigencia». En uso
  real, los metadatos caducan a los 300 s, precisamente porque son lo que dice si la norma sigue en vigor.
- **Qué faltó.** Que alguna eval lo mida. Las 17 de `boe-legislacion` leen artículos de normas vivas, así que ninguna
  comprueba que la skill **traslade el aviso a su respuesta**. `SKILL.md` dice que hay que señalarlo (FR-011, SC-004
  lo cuenta como regla presente en el texto), pero nadie comprueba que ocurra en una sesión real. Y el formato de eval
  tampoco lo permitiría hoy: solo sabe exigir comandos ejecutados y citas encontradas, no que la respuesta lleve un
  aviso.
- **Qué se hizo.** Anotarlo. Arreglarlo son tres piezas: un campo nuevo en el formato de eval (los avisos esperados),
  su comparación mecánica en `Juzgar` como la de las citas, y una eval que las use. La norma candidata ya está medio
  grabada: **`BOE-A-1992-26318`, la Ley 30/1992**, que H4 grabó para sus propios tests y cuyos metadatos dan
  `estatus_derogacion: "S"` y `vigencia_agotada: "S"` —o sea, dos avisos, `derogada` y `vigencia-agotada`—, con su
  bloque `a42` ya grabado. Falta grabar su `indice` (FR-074 lo exige de toda norma de una eval) y su `buscar`, si la
  norma entra en `data/normas.yaml`, que la regla «normas conocidas» obliga. Eso es la única pausa humana:
  `scripts/grabar-evals.sh`. No es un hito: cabe en una sesión.

- **Qué se pidió.** Cerrar H5 con las evals de `boe-legislacion` en verde. El job las ejecuta con
  `claude-haiku-4-5-20251001`, fijado por la clarificación Q5 del spec de H5 («un único modelo de gama
  económica») que viene de la tabla de controles del roadmap (`docs/ROADMAP.md` §4: «job semanal con modelo
  barato»). Al revisar las pausas del hito, Jorge preguntó si Haiku es el modelo adecuado para esto y para
  qué sirve la ejecución semanal.
- **Qué falló.** Nada del job ni de la skill, pero tres tareas de H5 existen solo para acomodar lo que el
  modelo hace de forma no fiable: T041 y T043 (las diez preguntas positivas nombran el artículo, porque el
  modelo no acierta el artículo por materia; entrada de abajo), T046 (la extracción de la cita admite la
  forma legible dentro de los corchetes, porque tras dos refuerzos del protocolo el modelo seguía metiéndola
  ahí en una sesión de cada diez). Las dos ejecuciones de cierre (35002104338 y 35023013878) dieron 10 de 10,
  pero «10 de 10 en una sola ejecución» es frágil con cualquier LLM. Y la ejecución semanal sobre `main`, con
  modelo, versión de Claude Code y respuestas del BOE fijados, mide sobre todo el azar del modelo: no hay
  cambios que comprobar, gasta suscripción y asume cada semana el riesgo del token en `/proc/<pid>/environ`
  (research D13 de H5).
- **Qué faltó.** Que el job decida con el modelo del uso real de la skill, que la medida no dependa de una
  sola tirada y que se ejecute cuando hay algo que medir.
- **Qué se hizo.** Decisión de Jorge (2026-09-15), pieza aparte tras fusionar H5 (#27), como ADR con enmienda
  de Q5, FR-070, SC-003 y research D13, y del roadmap:
  1. **Decide Sonnet 5** (`claude-sonnet-5`) [enmienda 2026-09-30, ADR 0031: decide el id al que resuelve `sonnet`
     en la versión de Claude Code del job, hoy `claude-sonnet-5-5`]: si funciona con Sonnet funciona con Opus, y gasta menos
     suscripción que Opus. La elección es de calidad, no de coste por token (la suscripción no se factura por
     llamada).
  2. **Haiku 4.5 queda informativo**: se ejecuta y se publica en el informe como límite inferior, pero no
     hace fallar el job.
  3. **Cada eval se repite** N veces (p. ej. 3) con umbral (≥ 2 de 3) y la tasa en el informe, en lugar de
     exigir 10 de 10 en una sola ejecución.
  4. **Vuelven las preguntas por materia**, sin nombrar el artículo, junto a las que lo nombran. La
     herramienta que las haría posibles sigue en el backlog (entrada de abajo).
  5. **La extracción tolerante de la cita (T046) se queda**: compara por identificador, que es lo que SC-009
     quiere medir.
  6. **Disparador por cambios en vez de semanal**: quitar el `schedule` (`cron '41 4 * * 1'`) de
     `.github/workflows/evals.yml` y lanzar el job en las propuestas de cambio que toquen `skills/`,
     `evals/`, `data/`, el applet `boe` o el propio job, además del lanzamiento manual y del de etiqueta.

### 2026-09-15 · `boe-legislacion`: encontrar un artículo por su materia dentro de una norma

- **Qué se pidió.** Preguntas por materia, sin el número del artículo, en las pruebas de red de las evals
  de H5 (job `evals`, sesiones con `claude-haiku-4-5-20251001`). En la ejecución 34941499481: «¿Qué impuestos
  pueden exigir los ayuntamientos según el texto refundido de la Ley reguladora de las Haciendas Locales?»
  (eval 05; se esperaba el artículo 59, `BOE-A-2004-4214` bloque `a59`) y «¿Qué dice la Ley 40/2015 sobre el
  principio de legalidad en la potestad sancionadora?» (eval 07; se esperaba el artículo 25,
  `BOE-A-2015-10566` bloque `a25`). En la ejecución 34956596912, tres que en la anterior habían salido bien:
  «¿Qué atribuciones tiene el Pleno del ayuntamiento según la Ley reguladora de las Bases del Régimen Local?»
  (eval 03; se esperaba el artículo 22, `BOE-A-1985-5392` bloque `a22`), «¿Qué rendimientos se consideran
  rendimientos íntegros del trabajo en la ley del IRPF?» (eval 06; se esperaba el artículo 17,
  `BOE-A-2006-20764` bloque `a17`) y «¿En qué plazo hay que resolver una solicitud de acceso a la información
  pública según la Ley 19/2013?» (eval 08; se esperaba el artículo 20, `BOE-A-2013-12887` bloque `a20`).
- **Qué falló.** La skill identificó la norma y leyó su índice (código 0), pero pidió otro artículo: en la
  05, `a2` (dos veces); en la 07, `a140` a `a145` con `articulos` y después `a140`; en la 03, `a21`
  (atribuciones del Alcalde); en la 06, `a21` (rendimientos del capital); en la 08, `a12` (derecho de
  acceso), dos veces. En el job solo responden los bloques grabados, así que esas lecturas terminaron con
  código 5 (la 03 obtuvo 4 al repetirla con `--offline`, y la 06, 2 con `--timeout 10000`) y las respuestas
  dijeron que no pudieron consultar la fuente, sin texto ni cita. La 03, la 06 y la 08 habían pedido el
  artículo esperado en la ejecución anterior: el número que recuerda el modelo cambia de una sesión a otra.
- **Qué faltó.** Una herramienta que, dentro de una norma, encuentre el artículo que trata una materia. El
  índice del BOE (`boe indice`) solo da «Artículo N», sin rúbrica, y las divisiones («TÍTULO I», «CAPÍTULO
  III»…): sirve para pasar del número de un artículo a su id, no para saber qué artículo regula qué. Hoy lo
  decide lo que el modelo sabe de memoria o, con red, leer bloques a tientas.
- **Qué se hizo.** Las preguntas de la 05 y la 07 nombran el artículo desde T041, y desde T043 lo nombran
  las diez positivas de las evals de `boe-legislacion` (H5), para que midan el protocolo de la skill y no la
  memoria del modelo. La herramienta queda en el backlog, como candidata del grupo «Profundidad del BOE y de
  los escritos» (`docs/ROADMAP.md` §4).
