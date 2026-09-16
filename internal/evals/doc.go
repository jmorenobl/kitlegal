// Package evals es la herramienta de desarrollo con la que el repositorio mide
// sus skills con evals: lee cada fichero de evals/<skill>/ en el formato común de
// schemas/eval.yaml.json y lo entrega como una Eval comparable por identificador
// (FR-060, FR-061; research.md D10).
//
// Lee cada eval con el lector común de documentos YAML de internal/skills, que
// rechaza toda clave repetida y valida contra el esquema, y todo error que
// devuelve empieza por el nombre del fichero (US4, escenario 4).
//
// El binario distribuido no enlaza este paquete: solo se ejecuta desde tests, y
// TestDependenciasDelBinario falla si algún día llegara al binario (research.md
// D2). Por eso lee el esquema del repositorio con una ruta relativa al directorio
// del paquete, que es donde go test ejecuta sus tests (research.md V46).
package evals
