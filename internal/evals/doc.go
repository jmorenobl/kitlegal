// Package evals es la herramienta de desarrollo con la que el repositorio mide
// sus skills con evals: lee cada fichero de evals/<skill>/ en el formato común de
// schemas/eval.yaml.json y lo entrega como una Eval comparable por identificador
// (FR-060, FR-061; research.md D10), también con las skills que la sesión no
// activa (no_se_activan) y las líneas ⚠ REDACCIÓN MODIFICADA: que la respuesta
// lleva (redacciones_modificadas; FR-003, FR-004 y FR-053 de H7.4), y con la
// eval sin binario ni servidor, la que declara sin_binario_ni_servidor: su
// sesión tiene la skill y nada con lo que consultar, y no lleva comandos ni nada
// de lo que una respuesta con consulta debe traer (FR-046 de H21). Lee también,
// si la carpeta la tiene, la lista de expresiones prohibidas de la skill,
// expresiones-prohibidas.yaml en el formato de
// schemas/expresiones-prohibidas.yaml.json, con sus cuatro familias de dos
// clases —de la A, la maquinaria, lo dicho en otra conversación y el anuncio de
// la respuesta; de la B, redaccion_no_leida, una redacción que ninguna orden
// devolvió— y sus formas_fijas, que ExtraerExpresionesProhibidas quita de la
// respuesta antes de buscar, y la deja en cada Eval de la carpeta (FR-050 de
// H7.2; FR-020 de H7.3; FR-030 y FR-031 de H7.4).
//
// Lee cada eval y la lista con el lector común de documentos YAML de
// internal/skills, que rechaza toda clave repetida y valida contra el esquema, y
// todo error que devuelve empieza por el nombre del fichero (US4, escenario 4).
//
// Con esas evals abre y juzga las sesiones de Claude Code del job de evals, sin
// ningún modelo en el juicio:
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
//     (FR-042 de H21);
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
//     ExtraerSinConsulta, y ninguna cita (FR-047 de H21);
//   - EscribirInforme escribe informe.json e informe.md con la tasa de cada
//     serie —las sesiones de una eval con un modelo en un modo—, el recuento de
//     expresiones prohibidas por modelo y modo, solo de las sesiones terminadas,
//     los umbrales con el contrato del ADR 0029, cada uno de un modo, que lo
//     nombra —los tres de la respuesta del modelo que decide,
//     expresiones_prohibidas, sin_activar y redaccion_no_leida, y el de la
//     duración de la tanda de cada modo hacen fallar el veredicto; los de los
//     demás modelos solo se publican; las sesiones de la eval sin binario ni
//     servidor no entran en ninguno—, las sesiones sin medir, los reintentos por
//     límite de ritmo y la duración, y da su veredicto: una serie que decide y no
//     llega al umbral en un modo lo pone en fallo aunque pase en el otro (FR-001
//     a FR-008, FR-033 y FR-051 de H7.3; FR-040 a FR-045 de H7.4; FR-043 a
//     FR-048 de H21);
//   - leerDefinicionDelJob lee la definición del job, .github/workflows/evals.yml:
//     con ella la comprueba TestDefinicionDelJob en make ci, y de ella toma el
//     sondeo la concurrencia de cada skill (FR-034, FR-035 y FR-094 de H7.3);
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
//     Un argumento que no vale —también el número de una eval sin binario ni
//     servidor— o la credencial que falta son un errorDeUso, que no abre ninguna
//     sesión y que quien lo lanza lee solo, sin la traza de go test (FR-080 de
//     H7.4).
//
// Los puntos de entrada que abren sesiones con modelo o consultan la plataforma
// llevan la etiqueta de construcción evals y quedan fuera de make ci:
// TestEjecucionDelJob, al que llama scripts/evals.sh en el job; TestTandaDelCommit,
// al que llama el trabajo tanda del job, que consulta con gh las ejecuciones del
// flujo sobre el commit y deja medir=si o medir=no en su salida; y TestSondeo, al
// que llama scripts/evals-sondeo.sh desde make evals-sondeo, que deja el mensaje
// de un errorDeUso en uso.txt para que el guion lo imprima sin nada más. En make
// ci, los tests sustituyen claude, strace y go por guiones que no abren ninguna
// sesión, y la decisión de la tanda se prueba con consultas sintéticas (research.md
// D18 de H7.3; FR-099 y FR-100 de H7.4).
//
// El binario distribuido no enlaza este paquete: solo se ejecuta desde tests, y
// TestDependenciasDelBinario falla si algún día llegara al binario (research.md
// D2). Por eso lee los esquemas del repositorio con una ruta relativa al
// directorio del paquete, que es donde go test ejecuta sus tests (research.md
// V46).
package evals
