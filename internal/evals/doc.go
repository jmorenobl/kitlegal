// Package evals es la herramienta de desarrollo con la que el repositorio mide
// sus skills con evals: lee cada fichero de evals/<skill>/ en el formato común de
// schemas/eval.yaml.json y lo entrega como una Eval comparable por identificador
// (FR-060, FR-061; research.md D10). Lee también, si la carpeta la tiene, la
// lista de expresiones prohibidas de la skill, expresiones-prohibidas.yaml en el
// formato de schemas/expresiones-prohibidas.yaml.json, con sus tres familias
// —la maquinaria, lo dicho en otra conversación y el anuncio de la respuesta—, y
// la deja en cada Eval de la carpeta (FR-050 de H7.2; FR-020 de H7.3).
//
// Lee cada eval y la lista con el lector común de documentos YAML de
// internal/skills, que rechaza toda clave repetida y valida contra el esquema, y
// todo error que devuelve empieza por el nombre del fichero (US4, escenario 4).
//
// Con esas evals abre y juzga las sesiones de Claude Code del job de evals, sin
// ningún modelo en el juicio:
//
//   - el repartidor, ejecutarSesiones, abre las sesiones de un plan en su orden,
//     como mucho Concurrencia a la vez, cada una preparada justo antes, en su
//     propio directorio y con su tope de 240 s, con scripts/evals-sesion.sh; tras
//     el mensaje del límite de uso de la cuenta no abre ninguna más, y mide la
//     duración de las sesiones (FR-030, FR-031, FR-044 y FR-050 de H7.3);
//   - Juzgar decide si pasa cada sesión con lo que dejan su transcript y su
//     traza, y ClasificarElLimite reconoce, en el transcript, la que un límite de
//     la cuenta no dejó terminar, que queda sin medir y no es una eval fallida
//     (FR-040 a FR-043 de H7.3);
//   - EscribirInforme escribe informe.json e informe.md con la tasa de cada
//     serie, el recuento de expresiones prohibidas por modelo, los umbrales con
//     el contrato del ADR 0029 —el de las expresiones prohibidas del modelo que
//     decide y el de la duración de las sesiones hacen fallar el veredicto; los
//     de los demás modelos solo se publican—, las sesiones sin medir, los
//     reintentos por límite de ritmo y la duración, y da su veredicto (FR-001 a
//     FR-008, FR-033 y FR-051 de H7.3);
//   - leerDefinicionDelJob lee la definición del job, .github/workflows/evals.yml:
//     con ella la comprueba TestDefinicionDelJob en make ci, y de ella toma el
//     sondeo la concurrencia de cada skill (FR-034, FR-035 y FR-094 de H7.3);
//   - el sondeo, sondear, abre con el mismo repartidor y sin traza unas evals de
//     una skill con un solo modelo, con el binario y las skills del árbol de
//     trabajo, y las juzga con el mismo código, sin lo que sale de la traza: no
//     es un veredicto (FR-060 a FR-066 de H7.3).
//
// Los puntos de entrada que abren sesiones con modelo llevan la etiqueta de
// construcción evals y quedan fuera de make ci: TestEjecucionDelJob, al que llama
// scripts/evals.sh en el job, y TestSondeo, al que llama scripts/evals-sondeo.sh
// desde make evals-sondeo. En make ci, los tests sustituyen claude, strace y go por
// guiones que no abren ninguna sesión (research.md D18 de H7.3).
//
// El binario distribuido no enlaza este paquete: solo se ejecuta desde tests, y
// TestDependenciasDelBinario falla si algún día llegara al binario (research.md
// D2). Por eso lee los esquemas del repositorio con una ruta relativa al
// directorio del paquete, que es donde go test ejecuta sus tests (research.md
// V46).
package evals
