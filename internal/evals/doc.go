// Package evals es la herramienta de desarrollo con la que el repositorio mide
// sus skills con evals: lee cada fichero de evals/<skill>/ en el formato común de
// schemas/eval.yaml.json y lo entrega como una Eval comparable por identificador
// (FR-060, FR-061; research.md D10), también con las skills que la sesión no
// activa (no_se_activan) y las líneas ⚠ REDACCIÓN MODIFICADA: que la respuesta
// lleva (redacciones_modificadas; FR-003, FR-004 y FR-053 de H7.4), y con la
// eval sin binario ni servidor, la que declara sin_binario_ni_servidor: su
// sesión tiene la skill y nada con lo que consultar, y no lleva comandos ni nada
// de lo que una respuesta con consulta debe traer (FR-046 de H21). Desde H23
// una eval puede declarar además sentencias —lo que la respuesta debe llevar, y
// lo que no, de las sentencias de las que habla— y, entre sus comandos, los dos
// del applet cita, preparar y cotejar; con sentencias puede no llevar comandos,
// citas ni territorio (FR-052 de H23).
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
//   - las sentencias (sentencias.go) se juzgan por su forma, sin ningún modelo.
//     De lo que una eval declara en sentencias, Juzgar compara con la respuesta
//     las citas de sentencia —un corchete con el ECLI y el ROJ, que lee
//     ExtraerCitasDeSentencia—, la línea ⚠ SENTENCIA NO COMPROBADA:, que
//     reconoce ExtraerNoComprobada, y las direcciones y las casillas, que la
//     respuesta lleva tal cual; y con la sesión, la dirección de búsqueda que
//     devolvió en ella un cita preparar con texto
//     (ExtraerDireccionesDeBusqueda). Cada cosa que falta o que sobra es un
//     motivo, y un comando de cita lo cumple una invocación de su verbo, con el
//     --roj o el --texto que pida, como orden o como herramienta. La salida de
//     una operación es cada sobre de kitlegal entre los textos de la sesión
//     (ExtraerSobres): una orden sin --json no da ninguno. Con ellos,
//     CitaSinDocumento da el hecho de la sesión que mide el umbral
//     cita_sin_documento: los ECLI de una cita que ningún cita cotejar de la
//     sesión leyó en una ficha (ExtraerECLICotejados) y los que, fuera de una
//     cita y de una línea que empieza por ⚠, no están en ningún sobre de la
//     sesión ni en la pregunta (FR-051, FR-060 y FR-061 de H23);
//   - el juez vota las respuestas (juez.go). De cada una compone un mensaje con
//     los textos de sus herramientas, la pregunta y la respuesta, y nada más
//     (mensajeDelVoto), y pide cada voto a un Votante, que en el job, en la
//     medida y en el sondeo abre con scripts/evals-voto.sh una sesión nueva de
//     Claude Code sin herramientas, con la rúbrica como instrucciones, el esquema
//     como forma de la respuesta y un tope de 35 s. Comprueba sin modelo que la
//     frase que cita un sí está en la respuesta (fraseEsta); un sí sin su frase
//     hace nulo el voto, que se repite una vez, y un voto que el tope corta se
//     pide otra vez, también una sola (research «Reparaciones del cierre» de
//     H25); y una clase que decide marca la respuesta solo con tres votos que
//     dicen sí con su frase: se piden por orden, y se deja de votar en cuanto
//     ninguna clase que decide sigue con todos los suyos en sí. En una que solo
//     se publica cuenta el primero. Un voto que no llega a darse —el tope corta
//     también su repetición, o su sesión, su salida o su forma fallan— deja la
//     respuesta sin juzgar. Los votos no cambian el juicio sin modelo de la
//     sesión (FR-001 a FR-014 de H24);
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
//     H24). La reconstrucción sabe, desde H25, lo que piden los casos de
//     jurisprudencia. El informe de un caso es el de un job o el de un sondeo,
//     el que lleva la clave sondeo: la pregunta de una sesión suya sale de
//     preguntas.json, que está junto a él, y sus textos son las órdenes de
//     kitlegal y las salidas que el propio informe guarda, sin repetir ninguna.
//     En el de un job, una orden del applet cita se repite con sus argumentos
//     y, si es cotejar sin --documento, con lo que queda del texto pegado en la
//     pregunta por la entrada estándar; si termina con otro código que el que
//     le da el informe, el caso no se resuelve. Y un caso derivado quita un
//     bloque de una norma, de los textos, o un texto —documento, fallo o
//     apartado-2—, la parte del texto pegado que sale de su pregunta: sin el
//     documento, ninguna orden cotejar da texto (FR-040 a FR-045 de H25);
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
//     sus umbrales son, por cada modo, sin_activar y el recuento de cada clase
//     del juez, <clase>:<modelo>:<modo> —en boe-legislacion, afirma_lo_no_leido
//     y cuenta_su_proceso; en jurisprudencia, afirma_lo_no_leido y
//     afirma_que_existe—; desde H26, el umbral conjunto de cada clase que
//     decide, <clase>:<modelo>, sin modo —en las dos skills,
//     afirma_lo_no_leido—, que junta las respuestas de todos los modos y cuyo
//     umbral, el de la clase en clases.yaml, es una tasa sobre todas ellas; los
//     dos de la medida versionada de cada clase que decide,
//     medida_del_juez:<clase>:defectos_sin_marcar y
//     medida_del_juez:<clase>:correctos_marcados; y, por cada modo,
//     duracion_de_las_sesiones, si el job da un objetivo, y duracion_del_juez.
//     Hacen fallar el veredicto todos menos los recuentos de las clases del
//     juez en cada modo, que solo se publican, con el umbral de su clase: de
//     una clase que decide, lo que decide es su conjunto, cuyo motivo nombra
//     cada respuesta marcada, por su sesión, con su modo y con sus frases; y la
//     que solo se publica no tiene conjunto. Fuera de él, las medidas de un
//     modo no se suman a las del otro.
//     Con una skill cuyas evals declaran sentencias, tenga juez o no, los
//     umbrales de cada modo llevan además cita_sin_documento, las respuestas
//     con una cita sin documento cotejado o con un ECLI sin origen, que decide
//     con 0 y cuyo motivo nombra cada respuesta que cuenta, por su sesión y
//     con sus ECLI. Desde H23.1, si alguna de sus evals pide el cotejo —activa
//     la skill y espera algún cita cotejar—, detrás va documento_sin_cotejar,
//     las respuestas de esas evals cuya sesión no ejecutó ningún cita cotejar
//     que consultara y terminara bien, sobre las de esas evals, que decide con
//     0 y cuyo motivo nombra cada sesión que cuenta. Desde H9, con alguna eval
//     de plazo, detrás va plazo_sin_fecha, las respuestas de esas evals sin la
//     línea ⚠ PLAZO: con la fecha esperada, y con alguna de cobertura,
//     cobertura_sin_declarar, las de esas evals sin la línea ⚠ NO CUBIERTO:
//     con la norma esperada o con una línea ⚠ PLAZO:, cada uno sobre las de
//     sus evals, que deciden con 0 y cuyo motivo nombra cada sesión que
//     cuenta; con ellas, la skill tiene también sin_activar, aunque no tenga
//     juez ni sentencias. Desde el ADR 0042, con alguna eval de vigencia,
//     detrás de cobertura_sin_declarar va vigencia_sin_declarar, las de esas
//     evals sin la línea ⚠ PLAZO SIN CALCULAR: con la norma esperada o con una
//     línea ⚠ PLAZO:, igual. Desde H20, con alguna eval que espera una redacción
//     fechada, detrás de todos ellos va redaccion_sin_fechar, las respuestas de
//     esas evals cuya sesión no leyó con boe articulo el bloque de la eval en su
//     fecha o que no llevan las líneas ⚠ REDACCIÓN A FECHA:, ⚠ REDACCIÓN DE
//     HOY: y ⚠ SIN REDACCIÓN VIGENTE: que la eval espera, sobre las de esas
//     evals, que decide con 0 y cuyo motivo nombra cada sesión que cuenta; con
//     ellas, la skill tiene también sin_activar. Con el plan del job, los dos
//     modos, boe-legislacion lleva quince, de los que deciden once;
//     jurisprudencia, quince, de los que deciden once; y legal-core, diez, que
//     deciden los diez: de cada modo, sin_activar, plazo_sin_fecha,
//     cobertura_sin_declarar, vigencia_sin_declarar y redaccion_sin_fechar. En
//     boe-legislacion, con
//     su objetivo: de cada modo, el del modo orden delante, sin_activar, los
//     recuentos de afirma_lo_no_leido y de cuenta_su_proceso y
//     redaccion_sin_fechar; el conjunto de afirma_lo_no_leido; los dos de la
//     medida versionada; y duracion_de_las_sesiones y duracion_del_juez de cada
//     modo. En
//     jurisprudencia, que desde H25 tiene juez y sigue sin objetivo: de cada
//     modo, sin_activar, los recuentos de afirma_lo_no_leido y de
//     afirma_que_existe, cita_sin_documento y documento_sin_cotejar; el
//     conjunto de afirma_lo_no_leido; los dos de la medida versionada, sobre
//     sus 125 defectos y sus 124 correctos; y duracion_del_juez de cada modo,
//     sin duracion_de_las_sesiones. En las dos, no deciden los cuatro
//     recuentos. Una skill sin juez, sentencias ni evals de plazo, de
//     cobertura o con una redacción fechada no tiene ninguno de ellos.
//     Desde H26.1, cada serie lleva sesiones_con_lectura_sin_servir: en la de
//     una eval con sin_servir, cuántas de sus sesiones leídas tienen alguna
//     invocación en lecturas_sin_servir, también 0, y null en las demás, sin
//     decidir nada. Y el informe de toda skill lleva, detrás de todos los
//     umbrales, por cada modo del plan y, en cada uno, por cada modelo, el que
//     decide delante, fuera_de_lo_grabado:<modelo>:<modo>: las lecturas del
//     fuera_de_lo_grabado de sus sesiones medidas —legibles y no sin medir por un
//     límite de uso, de cualquier eval de ese modo, activen o no la skill y
//     terminen o no; ninguna de una eval sin binario ni servidor—, que ya no
//     lleva las de sin_servir, sobre esas sesiones, también con medida 0; no
//     decide nunca, y ni da motivo ni cambia el veredicto. Detrás del de cada
//     modelo y modo, el de una skill con alguna eval con un comando de bloque
//     cuya lectura está en su sin_servir lleva
//     lectura_sin_intentar:<modelo>:<modo>: las sesiones medidas de esas evals
//     que no tienen en lecturas_sin_servir ninguna invocación que consulte el
//     bloque de cada uno de esos comandos, sobre esas sesiones, también con
//     medida 0; decide con 0 solo el del modelo que decide, y su motivo nombra
//     cada sesión que cuenta. Con el plan del job, los dos modos y los dos
//     modelos, legal-core lleva dieciocho, de los que deciden doce, y
//     boe-legislacion y jurisprudencia, diecinueve, de los que deciden los once
//     de antes. Ninguno mide las expresiones de la lista: desde H24 no hay
//     expresiones_prohibidas ni redaccion_no_leida, ni recuento de expresiones
//     por modelo. La clave juez lleva los votos y las frases de cada respuesta
//     con algún voto afirmativo y las que quedaron sin juzgar, que ponen el
//     veredicto en fallo por la ejecución, no por la skill (FR-001 a FR-008,
//     FR-033 y FR-051 de H7.3; FR-040 a FR-045 de H7.4; FR-043 a FR-048 de H21;
//     FR-030 a FR-037 y FR-060 a FR-062 de H24; FR-060 a FR-063 de H23;
//     FR-020 a FR-025 de H25; FR-001 a FR-005 y FR-010 a FR-012 de H26;
//     FR-001 a FR-007 de H23.1; FR-053 a FR-056 de H9; FR-054 a FR-056 de
//     H20; FR-015, FR-021, FR-022 y FR-033 de H26.1);
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
//     H7.4). Desde H30 decide con el modo de la ejecución, completa o
//     reutilizacion: no mide si una anterior mide el commit con un modo igual o
//     mayor, y completa es mayor; lo lee de las dos marcas del trabajo tanda de
//     cada una (leerLaTandaDeLaEjecucion), la de que mide el commit y la de que
//     mide la suite completa (FR-021 de H30);
//   - la elección de una propuesta (dependencias.go y eleccion.go; FR-001 a
//     FR-008 de H30). calcularDependencias da, sobre el árbol de la cabeza y el
//     grafo de paquetes de go list (leerGrafoDePaquetes), el conjunto de
//     dependencias de cada eval: su fichero; el SKILL.md y las referencias de su
//     skill y de cada skill cuyo nombre aparece en el SKILL.md de otra del
//     conjunto; el fichero de data/ del que sale cada referencia
//     (skills.DatosDeLaReferencia); de cada applet de sus tablas de comandos
//     (skills.AppletsDeLaTabla), su fichero de internal/app y los paquetes del
//     módulo que alcanza, con sus ficheros empotrados y sus módulos externos;
//     las raíces de grabaciones; y la carpeta del juez y los demás ficheros de
//     evals/<skill>/ que no son evals. loComun da lo que es de todas. Las rutas
//     que miden y las de lo común son listas fijas en el código (rutaQueMide y
//     esDeLoComun), sin leer el disco ni la definición del job. Con los árboles
//     de la base y de la cabeza (leerArbolDeLaBase y leerArbolDeLaCabeza),
//     calculoDeLaEleccion.elegir aplica a cada ruta que la propuesta cambia, las
//     dos de un renombrado, la primera regla que vale —no mide; go.mod y go.sum,
//     por los módulos que cambian; lo común; una descripción que cambia; un
//     directorio de skills/ o de evals/ que está en un solo árbol; el conjunto de
//     alguna eval; y, si no, sin asignar— y da las evals elegidas, cada una con
//     el motivo de más prioridad y todas las rutas que la eligen. Un conjunto que
//     no se puede calcular, o los ficheros de la propuesta que gh no da, las
//     eligen todas (FR-005 de H30);
//   - la huella de cada eval (huella.go; FR-009 de H30): la sha256 del texto
//     canónico de calculoDeLaHuella.textos, con el modelo que decide y la versión
//     de Claude Code de las sesiones, los del juez si la skill lo tiene
//     (fijadosDeLaHuella), go y toolchain, los módulos de su conjunto y de lo
//     común, el resto de go.mod y de go.sum (restoDeGo), la descripción de cada
//     skill y cada ruta de su conjunto, de lo común y sin asignar con la sha256
//     de su contenido. Cualquier cambio en una de esas cosas la cambia;
//   - la reutilización (reutilizacion.go y plataforma.go; FR-010 a FR-013 de
//     H30). buscarEnLasMedidasDeMain mira las medidas de main —las
//     ejecuciones del flujo terminadas y sin cancelar sobre la cabeza de una
//     propuesta fusionada, con el trabajo de cada skill terminado y la marca del
//     informe escrito en verde, de menos de vidaDeUnResultado—, de la más
//     reciente a la más antigua y como mucho medidasQueSeRecorren; lee el
//     informe de la skill del registro de su trabajo (informeDelRegistro), y de
//     cada eval todavía sin resultado toma su entrada de origen con la misma
//     huella, completa y vigente (resultadoReutilizable). Para en cuanto cada una
//     tiene resultado, en la primera lectura que falla o con plazoDeLaPlataforma
//     agotado, y lo que falta se mide. Lee la plataforma por
//     lecturasDeLaPlataforma: en el job, ghDeLaPlataforma, con órdenes de solo
//     lectura de texto constante y cada valor en una variable del entorno; en
//     los tests, una en memoria;
//   - la decisión de cada trabajo (decision.go y origen.go; FR-012 a FR-019 y
//     FR-022 de H30). decidirElTrabajo da, con el modo de la tanda, el origen de
//     las evals de la skill: en completa, todas medidas, con suite_completa y sin
//     leer la plataforma; en reutilizacion, las que elige la propuesta medidas
//     con su motivo y sus rutas, y cada una de las demás reutilizada con su
//     procedencia o medida con sin_resultado_de_main. Es el fichero de la
//     decisión (escribirElFicheroDeLaDecision), que scripts/evals.sh exige en
//     DECISION_DE_EVALS y TestEjecucionDelJob lee con -decision: ejecutarElJob
//     abre solo las sesiones de lo medido (PlanDeEvals.deLasMedidas), y el
//     informe lleva el bloque origen —el modo, cuántas evals midió y cuántas
//     reutilizó, las sesiones que abrió y, de cada eval, su huella, su motivo o
//     su procedencia y sus aportes a cada umbral de las respuestas—, la sección
//     «Medidas y reutilizadas» de informe.md, las series de lo reutilizado y,
//     junto a cada umbral al que aporta alguna eval reutilizada,
//     <umbral>:reutilizadas, que decide como el original salvo
//     <clase>:<modelo>:reutilizadas. Los campos de antes cuentan solo lo medido,
//     y la suite completa da el informe de antes con su origen (FR-023 y FR-026
//     de H30);
//   - el límite de la tasa (tasa.go; ADR 0045): el umbral <clase>:<modelo> se
//     cumple si las marcadas de las n respuestas medidas no pasan de k(n), el k
//     más estricto con P(Bin(n, pref) > k) ≤ 0,02, con pref derivada del umbral
//     de clases.yaml y de las respuestas que juzga la ejecución completa
//     (nuevoLimiteDeLaTasa y limiteDeLaTasa.para), con enteros y sin redondeo; y
//     compararConLaTasa publica el umbral de clases.yaml si la comparación con
//     él da k(n), y si no k(n)/n. Con todas las respuestas es el umbral de antes
//     (FR-024 de H30);
//   - la comprobación de la release (suite_completa.go; FR-025 de H30):
//     comprobarLaSuiteCompleta mira las ejecuciones del flujo sobre el commit
//     etiquetado, de la más reciente a la más antigua, y vale la primera con el
//     trabajo de cada skill con evals en el commit (evalsDelCommit) en success y,
//     en su informe, el modo completa, cada eval medida y ninguna reutilizada; si
//     ninguna vale, falla con el commit, el motivo de cada ejecución mirada y la
//     orden que lanza la suite completa sobre la etiqueta, gh workflow run
//     evals.yml --ref <etiqueta>;
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
// TestEjecucionDelJob, al que llama scripts/evals.sh en el job, desde H30 con el
// fichero de la decisión en -decision, y que llama a ejecutarElJob;
// TestTandaDelCommit, al que llama el trabajo tanda del job con el modo de la
// ejecución en -modo, que consulta con gh las ejecuciones del flujo sobre el
// commit y deja medir=si o medir=no y modo=<modo> en su salida;
// TestEleccionDeLaPropuesta, al que llama el trabajo cambios al abrir o reabrir
// una propuesta, que calcula la elección, la registra y deja elige=si o elige=no
// en su salida (escribirLaApertura); TestDecisionDelTrabajo, al que llama el paso
// «Decidir qué mide y qué reutiliza» de cada trabajo evals, que escribe el
// fichero de la decisión y registra cada eval, medida con su motivo o
// reutilizada con su procedencia; TestSuiteCompletaDelCommit, al que llama el
// trabajo suite-completa del flujo release, que llama a comprobarLaSuiteCompleta
// con el commit etiquetado y su etiqueta; TestMedidaDelJuez, al que llama
// scripts/evals-medir-juez.sh desde make evals-medir-juez, que llama a
// medirAlJuez y escribe la medida para que el guion la imprima, y que lanza una
// persona, con la etiqueta evals-medir-juez o la entrada medir_al_juez del flujo,
// nunca un paso de un run; y TestSondeo, al que llama scripts/evals-sondeo.sh
// desde make evals-sondeo, que deja el mensaje de un errorDeUso en uso.txt para
// que el guion lo imprima sin nada más. En make ci, los tests sustituyen claude,
// strace y go por guiones que no abren ninguna sesión, los votos del juez son
// salidas grabadas, y la decisión de la tanda se prueba con consultas sintéticas
// (research.md D18 de H7.3; FR-099 y FR-100 de H7.4; FR-093 de H24). Desde H30,
// la elección, la huella y los conjuntos se calculan sobre el repositorio y sus
// mutantes en memoria (TestEleccionDeLosMutantes, TestMatrizDeLasPropuestas,
// TestHuellaDeLasEvals y TestConjuntoDeCadaEval), y la búsqueda de medidas de
// main, la decisión de cada trabajo y la comprobación de la release, con la
// plataforma en memoria (TestBusquedaDeMedidasDeMain, TestDecisionDelJob y
// TestComprobarLaSuiteCompleta); el límite de la tasa lo comprueba
// TestLimiteDeLaTasa para cada skill con juez y cada número de respuestas.
//
// El binario distribuido no enlaza este paquete: solo se ejecuta desde tests, y
// TestDependenciasDelBinario falla si algún día llegara al binario (research.md
// D2). Por eso lee los esquemas del repositorio con una ruta relativa al
// directorio del paquete, que es donde go test ejecuta sus tests (research.md
// V46).
package evals
