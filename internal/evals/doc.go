// Package evals es la herramienta de desarrollo con la que el repositorio mide
// sus skills con evals: lee cada fichero de evals/<skill>/ en el formato común de
// schemas/eval.yaml.json y lo entrega como una Eval comparable por identificador
// (FR-060, FR-061; research.md D10), también con las skills que la sesión no
// activa (no_se_activan) y las líneas ⚠ REDACCIÓN MODIFICADA: que la respuesta
// lleva (redacciones_modificadas; FR-003, FR-004 y FR-053 de H7.4), y con la
// eval sin binario ni servidor, la que declara sin_binario_ni_servidor: su
// sesión tiene la skill y nada con lo que consultar, y no lleva comandos ni nada
// de lo que una respuesta con consulta debe traer (FR-046 de H21).
//
// De la misma carpeta lee LeerConjunto, si la skill las tiene, dos cosas que no
// son evals:
//
//   - la lista de expresiones, expresiones-prohibidas.yaml en el formato de
//     schemas/expresiones-prohibidas.yaml.json: el vocabulario que la prosa del
//     SKILL.md de la skill no usa. Desde H24 no juzga ninguna respuesta (FR-070
//     y FR-071 de H24). Conserva las cuatro familias con las que nació —la
//     maquinaria, lo dicho en otra conversación, el anuncio de la respuesta y
//     redaccion_no_leida, una redacción que ninguna orden devolvió—, que
//     ExtraerExpresionesProhibidas busca en un texto después de quitarle sus
//     formas_fijas (FR-050 de H7.2; FR-020 de H7.3; FR-030 y FR-031 de H7.4), y
//     tiene además salida_de_las_herramientas, lo que devuelve una orden o una
//     herramienta y los nombres de sus campos, que en la prosa se busca también
//     dentro del código en línea (FR-085 de H24);
//   - el juez con modelo, la carpeta juez (juez.go; ADR 0037): sus clases,
//     clases.yaml en el formato de schemas/juez-clases.yaml.json, cada una con
//     si decide o solo se publica y con su umbral, y la rúbrica, el esquema de
//     su respuesta, los casos etiquetados y la medida versionada, que son copias
//     de la evidencia del ADR. Una skill sin esa carpeta no tiene juez (FR-020 a
//     FR-023 de H24).
//
// Lee cada eval, la lista y las clases con el lector común de documentos YAML de
// internal/skills, que rechaza toda clave repetida y valida contra el esquema, y
// todo error que devuelve empieza por el nombre del fichero (US4, escenario 4).
//
// Con esas evals abre y juzga las sesiones de Claude Code del job de evals. Lo
// que tiene forma o es un hecho de la sesión —citas, avisos, formas fijas,
// órdenes, activación, red, duración— se juzga sin ningún modelo, y nada de eso
// pasa al juez; lo que la respuesta quiere decir lo juzga el juez de la skill,
// si lo tiene:
//
//   - el plan, PlanDeEvals, da las sesiones de cada eval en cada uno de sus
//     modos, que son lo que la sesión tiene para consultar: en el modo orden,
//     ModoOrden, kitlegal en el PATH y ningún servidor; en el modo herramienta,
//     ModoHerramienta, el servidor MCP de kitlegal declarado en el servidor.json
//     de la sesión y kitlegal fuera del PATH. El job da los dos; sin ninguno, el
//     plan es solo del modo orden, que es el del sondeo. Las de la eval sin
//     binario ni servidor van una sola vez y sin modo, sin kitlegal en el PATH y
//     sin servidor.json. Sesiones las deja en tres tandas: las del modo orden,
//     las del modo herramienta y las de la eval sin binario ni servidor (FR-040,
//     FR-041 y FR-046 de H21);
//   - el repartidor, ejecutarSesiones, abre las sesiones de un plan en su orden,
//     como mucho Concurrencia a la vez, cada una preparada justo antes, en su
//     propio directorio y con su tope de 240 s, con scripts/evals-sesion.sh, que
//     declara el servidor con --mcp-config si la sesión tiene servidor.json; tras
//     el mensaje del límite de uso de la cuenta no abre ninguna más, y mide la
//     duración de las sesiones. El job lo llama una vez por tanda, con
//     ejecutarPorTandas, y mide cada una (FR-030, FR-031, FR-044 y FR-050 de
//     H7.3; FR-043 de H21);
//   - LeerSesion lee lo que deja cada sesión, y su respuesta es la de la
//     pregunta, el primer result del transcript: el aviso de una tarea en segundo
//     plano abre otro turno, con su propio result, que no la sustituye (FR-060 de
//     H7.4). Lee también sus llamadas, Sesion.Llamadas: cada tool_use del
//     transcript cuyo nombre, o lo que sigue a su último «__», es el de una
//     herramienta del registro de producción, con su resultado si lo tiene
//     (FR-042 de H21). Y los textos que devolvieron sus herramientas,
//     Sesion.Textos: lo que devolvió cada orden de Bash que nombra kitlegal y
//     cada una de esas llamadas, en su orden, que es lo que el juez recibe de la
//     sesión (FR-001 y FR-002 de H24);
//   - Juzgar decide si pasa cada sesión con lo que dejan su transcript y su
//     traza —también si activa una skill de no_se_activan y si falta la línea de
//     una redacción esperada, que ExtraerRedaccionesModificadas lee con su cita y
//     sus dos fechas—, y ClasificarElLimite reconoce, en el transcript, la que un
//     límite de la cuenta no dejó terminar, que queda sin medir y no es una eval
//     fallida (FR-040 a FR-043 de H7.3; FR-052 de H7.4). Las llamadas cuentan en
//     el juicio como invocaciones, con las reglas de la orden: el applet y el
//     verbo de la herramienta, los argumentos de cli.LineaDeLlamada y el código
//     de su clase, de modo que cumplen un comando esperado, incumplen uno
//     prohibido o quedan fuera de lo grabado igual que una orden; mcp serve en la
//     traza, el proceso del servidor, no es una consulta; y en una sesión sin
//     kitlegal en el PATH, una orden de kitlegal de otro applet la deja sin pasar
//     (FR-042 de H21). La de una eval sin binario ni servidor pasa si su
//     respuesta lleva la línea ⚠ SIN CONSULTA AL BOE: con su dirección, que lee
//     ExtraerSinConsulta, y ninguna cita (FR-047 de H21). No mira la lista de
//     expresiones: ninguna deja una sesión sin pasar (FR-070 de H24);
//   - el juez vota las respuestas (juez.go). De cada una compone un mensaje con
//     los textos de sus herramientas, la pregunta y la respuesta, y nada más
//     (mensajeDelVoto), y pide cada voto a un Votante, que en el job, en la
//     medida y en el sondeo abre con scripts/evals-voto.sh una sesión nueva de
//     Claude Code sin herramientas, con la rúbrica como instrucciones, el esquema
//     como forma de la respuesta y un tope de 35 s. Comprueba sin modelo que la
//     frase que cita un sí está en la respuesta (fraseEsta); un sí sin su frase
//     hace nulo el voto, que se repite una vez; y una clase que decide marca la
//     respuesta solo con tres votos que dicen sí con su frase: se piden por
//     orden, y se deja de votar en cuanto ninguna clase que decide sigue con
//     todos los suyos en sí. En una que solo se publica cuenta el primero. Un
//     voto que no llega a darse deja la respuesta sin juzgar. Los votos no
//     cambian el juicio sin modelo de la sesión (FR-001 a FR-014 de H24);
//   - la medida del juez (medida.go). comprobarLaMedida dice, sin modelo, si la
//     medida versionada corresponde a la rúbrica, a los casos, al modelo del juez
//     y a la versión de Claude Code de sus votos, y si se cumple, con ningún
//     defecto sin marcar y ningún correcto marcado: lo comprueban make ci
//     (TestMedidaVersionada, con lo que fija la definición del job), el job antes
//     de abrir ninguna sesión y el sondeo, y TestCopiasDelJuez, que las copias de
//     la carpeta son las de la evidencia. medirAlJuez repite la medida:
//     reconstruye en proceso, sin red y sin modelo, los textos de cada caso
//     etiquetado desde su informe versionado, lo vota con la misma regla y da la
//     medida de lo que hay, o falla con los casos que no dan lo que dice su
//     etiqueta o que quedan sin juzgar. No abre ninguna sesión de evals ni
//     escribe en el repositorio: la versiona una persona (FR-040 a FR-054 de
//     H24);
//   - el recorrido del job (ejecucion.go), ejecutarElJob, va del plan al
//     informe: con una skill que tiene juez comprueba antes de nada su medida, y
//     si no corresponde o no se cumple escribe el informe del instrumento sin
//     medir sin abrir ninguna sesión, ni de evals ni del juez; si no, abre las
//     sesiones tanda a tanda y escribe el informe, sin votar ningún caso
//     etiquetado (FR-043 y FR-044 de H24);
//   - EscribirInforme escribe informe.json e informe.md con la tasa de cada
//     serie —las sesiones de una eval con un modelo en un modo—, los umbrales con
//     el contrato del ADR 0029, la clave juez, las sesiones sin medir, los
//     reintentos por límite de ritmo y la duración, y da su veredicto: una serie
//     que decide y no llega al umbral en un modo lo pone en fallo aunque pase en
//     el otro. Con una skill que tiene juez juzga antes las respuestas del modelo
//     que decide —las de cada modo en las evals que activan la skill y, aparte,
//     las de la eval sin binario ni servidor, que no entran en ningún umbral—, y
//     sus umbrales son, por cada modo, sin_activar y el de cada clase del juez
//     —en boe-legislacion, afirma_lo_no_leido, que decide, y cuenta_su_proceso,
//     que solo se publica—; los dos de la medida versionada de cada clase que
//     decide, medida_del_juez:<clase>:defectos_sin_marcar y
//     medida_del_juez:<clase>:correctos_marcados; y, por cada modo,
//     duracion_de_las_sesiones, si el job da un objetivo, y duracion_del_juez.
//     Hacen fallar el veredicto todos menos el de la clase que solo se publica.
//     Ninguno mide las expresiones de la lista: desde H24 no hay
//     expresiones_prohibidas ni redaccion_no_leida, ni recuento de expresiones
//     por modelo. La clave juez lleva los votos y las frases de cada respuesta
//     con algún voto afirmativo y las que quedaron sin juzgar, que ponen el
//     veredicto en fallo por la ejecución, no por la skill (FR-001 a FR-008,
//     FR-033 y FR-051 de H7.3; FR-040 a FR-045 de H7.4; FR-043 a FR-048 de H21;
//     FR-030 a FR-037 y FR-060 a FR-062 de H24);
//   - leerDefinicionDelJob lee la definición del job, .github/workflows/evals.yml:
//     con ella la comprueba TestDefinicionDelJob en make ci —también el modelo
//     del juez y la versión de Claude Code de sus votos, fijados aparte de los de
//     las sesiones, el trabajo de la medida y los dos topes—, y de ella toma el
//     sondeo la concurrencia de cada skill y el modelo del juez (FR-034, FR-035 y
//     FR-094 de H7.3; FR-090 a FR-092 y FR-112 de H24);
//   - la tanda, esperarLaDecision, decide si una ejecución del flujo mide su
//     commit: no lo mide si una anterior sobre el mismo commit, sin terminar, ya
//     lo mide (decidirLaTanda), y espera a que decidan las anteriores, nunca a su
//     tanda, consultando cada esperaEntreConsultas, como mucho
//     esperaMaximaDeLaDecision; así un segundo disparo no abre ninguna sesión, y
//     la etiqueta sobre un commit ya medido vuelve a medir (FR-070 y FR-071 de
//     H7.4);
//   - el sondeo, sondear, abre con el mismo repartidor y sin traza unas evals de
//     una skill con un solo modelo, con el binario y las skills del árbol de
//     trabajo, y las juzga con el mismo código, sin lo que sale de la traza: no
//     es un veredicto (FR-060 a FR-066 de H7.3). Mide solo el modo orden: el modo
//     herramienta y la eval sin binario ni servidor son del job (FR-050 de H21).
//     Con una skill que tiene juez, juzga además sus respuestas con él, con el
//     modelo del juez de la definición del job y el claude del equipo de quien lo
//     lanza, y su salida da, donde iba el recuento de expresiones, las marcadas
//     en cada clase, si la medida versionada corresponde a ese Claude Code y las
//     respuestas sin juzgar; no falla ni deja de juzgar por ninguna de las tres
//     cosas (FR-075 y FR-076 de H24). Un argumento que no vale —también el número
//     de una eval sin binario ni servidor— o la credencial que falta son un
//     errorDeUso, que no abre ninguna sesión y que quien lo lanza lee solo, sin
//     la traza de go test (FR-080 de H7.4).
//
// Los puntos de entrada que abren sesiones con modelo o consultan la plataforma
// llevan la etiqueta de construcción evals y quedan fuera de make ci:
// TestEjecucionDelJob, al que llama scripts/evals.sh en el job y que llama a
// ejecutarElJob; TestTandaDelCommit, al que llama el trabajo tanda del job, que
// consulta con gh las ejecuciones del flujo sobre el commit y deja medir=si o
// medir=no en su salida; TestMedidaDelJuez, al que llama
// scripts/evals-medir-juez.sh desde make evals-medir-juez, que llama a
// medirAlJuez y escribe la medida para que el guion la imprima, y que lanza una
// persona, con la etiqueta evals-medir-juez o la entrada medir_al_juez del flujo,
// nunca un paso de un run; y TestSondeo, al que llama scripts/evals-sondeo.sh
// desde make evals-sondeo, que deja el mensaje de un errorDeUso en uso.txt para
// que el guion lo imprima sin nada más. En make ci, los tests sustituyen claude,
// strace y go por guiones que no abren ninguna sesión, los votos del juez son
// salidas grabadas, y la decisión de la tanda se prueba con consultas sintéticas
// (research.md D18 de H7.3; FR-099 y FR-100 de H7.4; FR-093 de H24).
//
// El binario distribuido no enlaza este paquete: solo se ejecuta desde tests, y
// TestDependenciasDelBinario falla si algún día llegara al binario (research.md
// D2). Por eso lee los esquemas del repositorio con una ruta relativa al
// directorio del paquete, que es donde go test ejecuta sus tests (research.md
// V46).
package evals
