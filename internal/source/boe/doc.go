// Package boe es el adaptador de la fuente boe.legislacion-consolidada, la API
// de Legislación Consolidada del BOE: implementa el puerto core.Source para los
// seis verbos del applet boe —buscar, indice, articulo, articulos, metadatos y
// analisis— y es el porte a Go de refs/boe.py, la copia congelada de la skill
// boe-fiscal (FR-120, FR-121).
//
// El porte promete el contenido de boe.py, no su salida: los mismos recursos
// pedidos con los mismos parámetros, la misma versión vigente, el mismo texto,
// las mismas condiciones de aviso y las mismas vigencias de caché. La
// presentación la hereda el applet del kernel, y los fallos terminan con los
// códigos de salida estables y el sobre obligatorio. Donde el porte se aparta de
// boe.py, el requisito del spec de H4 lo dice con su motivo y la entrada de la
// sección siguiente lo anota.
//
// # Comportamientos no obvios de refs/boe.py
//
// refs/boe.py se leyó línea a línea antes de portar. Cada entrada anota un
// comportamiento con sus líneas en ese fichero, su destino —se porta, se adapta
// y por qué, o no se porta y por qué— y el requisito del spec de H4 que lo fija,
// o «Fuera de alcance»:
//
//   - [1] líneas 37-46 · Las vigencias de caché por verbo: 300 s para buscar y metadatos, y 7 días para indice, articulo y analisis · se porta · FR-091
//   - [2] líneas 411-416 y 433 · Los avisos de vigencia se guardan dentro de la entrada del artículo y se sirven con ella durante sus 7 días, aunque los metadatos de los que salen solo duren 300 s · se porta · FR-091
//   - [3] líneas 56 y 66 · Cada entrada guarda el instante ts en que se escribió, y boe.py solo lo usa para saber si ha caducado · se porta: para la vigencia y, además, como la fecha_consulta de lo que se sirve desde la caché · FR-096
//   - [4] líneas 49-66 · La caché en disco, un fichero JSON por entrada nombrado con el MD5 del verbo y de los argumentos unidos por «|», de modo que dos consultas con argumentos distintos pueden compartir entrada · se adapta: la caché es la de internal/cache y la clave la fija la fuente con el verbo y todo argumento que cambie la respuesta, sin que dos consultas distintas compartan entrada · FR-090
//   - [5] líneas 110-118 · La versión vigente es la última del bloque: su fecha es fecha_publicacion, aunque venga vacía, u «original» si falta el atributo, y fecha_vigencia e id_norma quedan vacías si faltan; un bloque sin versiones da el texto de todo el bloque, la fecha «original» y ni fecha de vigencia ni norma modificadora · se porta · FR-011
//   - [6] líneas 132-136 · El texto es el de ET.tostring en modo texto, que incluye el que sigue al cierre del elemento, con cada línea recortada de espacio en blanco, sin las líneas vacías y unidas por salto de línea · se porta · FR-011
//   - [7] líneas 99-100 y 139-150 · Un XML ilegible pasa a _fallback_parse, que saca el título y los párrafos con expresiones regulares y, si no halla ninguno, devuelve el cuerpo recortado a 2000 caracteres · no se porta: entregaría como texto legal un mensaje de error, un HTML o un fragmento sin título ni fecha; el porte falla con la clase fuente-no-disponible · FR-014
//   - [8] líneas 102-104 · Un XML legible sin ningún elemento bloque por debajo de la raíz, porque find no mira la raíz misma, devuelve el cuerpo recortado a 2000 caracteres con título y fecha «?» · no se porta: ese cuerpo no es un artículo y no se presenta como tal; el porte falla con la clase fuente-no-disponible · FR-014
//   - [9] líneas 104, 150, 289-296, 387, 420, 473-482, 533, 559-561 y 579-581 · El marcador «?» en lugar de un campo que la fuente no da · se adapta: el campo queda ausente o vacío según el esquema, porque «?» es un recurso de la presentación en texto plano, no se distingue de un valor real y rompería los tipos del esquema · FR-016
//   - [10] líneas 186-193 · _check_vigencia devuelve None, sin aviso ninguno, cuando los metadatos no se obtienen, no se interpretan o llegan vacíos, y el artículo sale sin advertencias · no se porta: callaría un fallo y presentaría sin advertencia un texto que puede estar derogado; articulo falla con la clase de ese fallo y no emite el texto · FR-013
//   - [11] línea 186 · _check_vigencia pide los metadatos sin leer ni escribir la caché, aunque cmd_metadatos guarda ese mismo recurso 300 s · se adapta: articulo y articulos leen y escriben la entrada de metadatos de la norma, porque es el mismo recurso con la misma vigencia y pedirlo otra vez solo añade carga a la fuente · FR-090
//   - [12] líneas 405-411 · cmd_articulo pide los metadatos aunque el bloque haya fallado, porque el error del bloque le llega como texto · se adapta: si el bloque falla, articulo falla con la clase de ese fallo sin pedir los metadatos, una petición cuyo resultado no se entregaría · FR-013
//   - [13] líneas 437-443 · cmd_articulos llama a cmd_articulo por cada bloque y, como _check_vigencia no usa la caché, repite la petición de metadatos por cada bloque que no está en caché · se adapta: los metadatos se piden como mucho una vez por invocación y un id repetido se pide una sola vez, para no pedir al BOE lo que ya se obtuvo · FR-020
//   - [14] línea 433 · cmd_articulo guarda cada bloque, con sus avisos, en cuanto lo obtiene, de modo que cmd_articulos deja en caché cada bloque por separado · se porta · FR-021
//   - [15] líneas 440-442 · cmd_articulos sigue pidiendo los bloques posteriores a uno fallido y los entrega todos, el fallido como texto · se adapta: articulos se detiene en el primer bloque que falla y falla nombrándolo, sin pedir los posteriores ni entregar los demás a medias · FR-021
//   - [16] líneas 99-104, 139-150, 186-193, 405-416 y 433 · La caché de fallos, que solo hace cmd_articulo: guarda 7 días el cuerpo del error o su recorte como si fuera el artículo, y el artículo sin avisos cuando los metadatos fallan · no se porta: serviría un error o un artículo sin avisos como respuesta válida, también con --offline; solo se guarda lo que la fuente respondió y se interpretó · FR-093
//   - [17] líneas 485-498 · cmd_metadatos redacta los avisos a su manera: sangrados, «la consolidación no está finalizada» en lugar de «la consolidación de esta norma no está finalizada», y «⚠ NORMA DEROGADA» y «⚠ VIGENCIA AGOTADA» sin la frase explicativa · se adapta: metadatos usa los textos de _check_vigencia, con las mismas condiciones y el mismo orden, para que cada codigo de aviso tenga una sola frase · FR-050
//   - [18] líneas 282-283 y 302 · Una búsqueda sin resultados retorna antes de guardarse, mientras que la que tiene resultados se guarda, así que la misma búsqueda vacía se vuelve a pedir · se adapta: una búsqueda sin resultados es un resultado, termina con código 0 y lista vacía y se guarda 300 s como cualquier otro · FR-032
//   - [19] líneas 362-363, 459-460 y 518-519 · Un índice, unos metadatos o un análisis vacíos retornan antes de guardarse · se porta: son «no encontrado», con código 3, y un fallo no se guarda · FR-041, FR-051, FR-061 y FR-093
//   - [20] líneas 218-225 · _ensure_list convierte en lista de un elemento el objeto suelto que la API a veces entrega en lugar de una lista, pero boe.py solo la aplica en sumario · se adapta: el porte la aplica en los seis verbos, porque el recorrido por claves que boe.py hace en buscar, indice y analisis pierde o desfigura resultados, bloques, materias o referencias que la fuente sí entrega · FR-070
//   - [21] líneas 195, 463, 522, 531, 540, 550-554 y 570-574 · La lectura de los envoltorios: el primer elemento de data o el objeto suelto en metadatos y análisis (195, 463 y 522), cada materia con envoltorio materia o directa (531), las notas con envoltorio nota o directas (540) y cada referencia con envoltorio anterior o posterior o directa (550-554 y 570-574) · se porta · FR-070
//   - [22] líneas 365-369 · El índice llega anidado bajo bloque en el primer elemento de data, o plano en data, y los dos se leen igual · se porta · FR-040
//   - [23] línea 107 · En cmd_articulo, el tipo del bloque es su atributo tipo en la respuesta XML, no el que _tipo_from_id infiere del id · se porta · FR-010
//   - [24] líneas 155-176 y 389 · En cmd_indice, el tipo de cada bloque se infiere de su id con _tipo_from_id, y un id sin regla da tipo vacío · se porta · FR-040
//   - [25] líneas 158-175 y 375 · Las rarezas de _tipo_from_id: decide el orden de las comprobaciones, así que todo id que empieza por t es titulo, y ci y cv ya los cubre la expresión ^c[ivxlcdm]; y la tabla de sangrado lleva subseccion, que ninguna regla produce · se porta: las reglas en su orden; la tabla de sangrado no, porque solo sirve a la presentación en texto plano · FR-040 y Fuera de alcance
//   - [26] líneas 252, 258-271 y 784 · buscar une los argumentos con un espacio y envía el texto tal cual si contiene « AND », « OR », « NOT », titulo:, materia: o una comilla doble; si no, exige cada palabra en el título como titulo:<palabra>, unidas con « AND »; el cuerpo es el json.dumps de la consulta, que escapa lo no ASCII como \uXXXX, la dirección lo codifica con quote, que deja la barra sin codificar, y pide como mucho 10 resultados · se porta · FR-030
//   - [27] líneas 264-271 y 783 · buscar admite un texto vacío o solo de espacio en blanco, que da cero palabras y pide titulo: seguido de ese texto, y con una sola palabra envía titulo: seguido del texto sin recortar · se adapta: un texto sin ninguna palabra da código 2 antes de pedir, porque no busca nada, y una sola palabra va como titulo:<palabra>, porque el espacio en blanco que la rodea no forma parte de lo buscado · FR-030
//   - [28] líneas 186, 355, 406, 452 y 511 · Los ids de norma y de bloque se concatenan en la ruta de la petición sin validarlos · se adapta: se validan antes de cualquier petición y lo que no tiene su forma da código 2, para que ninguna entrada altere la petición ni gaste una petición al BOE · FR-080 y FR-081
//   - [29] línea 564 · El texto de cada referencia anterior se recorta a 200 caracteres · se adapta: va completo, porque el recorte sirve a la presentación en texto plano y en data presentaría como texto de la referencia un fragmento que la fuente no entrega así · FR-060
//   - [30] líneas 77-86, 276-280, 360, 363, 457, 460, 516, 519, 777-779 y 804-806 · Los errores de la fuente se devuelven como texto con código 0, y los de uso, con la ayuda y código 1 · no se portan: cada fallo termina con su código de salida estable y el sobre de fallo, que el proyecto ya tiene decidido y que permite a un agente distinguir un error del contenido legal · FR-100
//   - [31] líneas 32 y 81 · Las peticiones salen por robust_request, del módulo network_utils, con 30 s de espera · no se porta: los reintentos, el ritmo por sitio y robots.txt los aplica internal/httpx · FR-122
//   - [32] líneas 306-346, 588-689 y 692-732 · Los verbos buscar-materia, sumario y materias · no se portan: el hito no los incluye; sumario está en el backlog y lo fiscal, en la vertical boe-fiscal · Fuera de alcance
//
// # Supuestos sobre Python que no se comprueban en local
//
// El porte reproduce comportamientos de la biblioteca de CPython que boe.py usa y
// que solo se comprobarían ejecutando Python o leyendo la fuente C de
// _elementtree, que no está en local (research.md D20, S5). Los tests fijan lo
// portado y, si una grabación trae comentarios, las referencias del diff de
// aceptación, que escribe una persona, lo contrastan (FR-116):
//
//   - (a) str.split y str.strip tratan U+001C a U+001F como espacio en blanco,
//     además de los caracteres White_Space (líneas 135 y 264).
//   - (b) isdigit y \d solo difieren de unicode.IsDigit en dígitos no decimales,
//     que quedan fuera del alfabeto de los ids (líneas 158 y 164).
//   - (c) XMLParser.feed con una cadena ignora la codificación que declara el
//     documento (línea 98).
//   - (d) expat normaliza el valor de los atributos como manda XML 1.0 §3.3.3:
//     tabulador, salto y retorno pasan a espacio (líneas 106-107 y 116-118).
//   - (e) El acelerador C _elementtree, que ET.fromstring usa cuando existe,
//     trata comentarios e instrucciones de proceso como el TreeBuilder de Python
//     puro: no los inserta y deja unido el texto de sus dos lados (líneas 98 y
//     134).
//   - (f) expat rechaza un documento con más de un elemento raíz o con texto que
//     no sea espacio en blanco fuera de él (línea 98).
//
// Este fichero no contiene ninguna declaración, solo este comentario, y
// doc_test.go es su control: TestDocAnotaElPorte lo lee como go doc y exige las
// 32 entradas de FR-120 con las líneas, el destino y el requisito que les fija
// (SC-010).
package boe
