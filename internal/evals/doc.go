// Package evals es la herramienta de desarrollo con la que el repositorio mide
// sus skills con evals: lee cada fichero de evals/<skill>/ en el formato común de
// schemas/eval.yaml.json y lo entrega como una Eval comparable por identificador
// (FR-060, FR-061; research.md D10). Lee también, si la carpeta la tiene, la
// lista de expresiones prohibidas de la skill, expresiones-prohibidas.yaml en el
// formato de schemas/expresiones-prohibidas.yaml.json, y la deja en cada Eval de
// la carpeta (FR-050 de H7.2).
//
// Lee cada eval y la lista con el lector común de documentos YAML de
// internal/skills, que rechaza toda clave repetida y valida contra el esquema, y
// todo error que devuelve empieza por el nombre del fichero (US4, escenario 4).
//
// El binario distribuido no enlaza este paquete: solo se ejecuta desde tests, y
// TestDependenciasDelBinario falla si algún día llegara al binario (research.md
// D2). Por eso lee los esquemas del repositorio con una ruta relativa al
// directorio del paquete, que es donde go test ejecuta sus tests (research.md
// V46).
package evals
