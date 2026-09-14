// Package skills es la herramienta de desarrollo con la que el repositorio lee y
// comprueba lo que forma sus skills: los documentos YAML de datos, el frontmatter
// de SKILL.md y lo que se genera a partir de ellos (research.md D2).
//
// Su pieza común es el lector de documentos YAML de esquemas.go, el mismo para
// los datos, las evals y el frontmatter (data-model, «Lectura de documentos
// YAML»; research.md D8): rechaza toda clave repetida nombrando el mapa, la clave
// y sus dos líneas, normaliza el documento a tipos JSON y lo valida contra su
// esquema presentando cada incumplimiento con la ruta dentro del documento y su
// línea.
//
// El binario distribuido no enlaza este paquete: solo se ejecuta desde tests, y
// TestDependenciasDelBinario falla si algún día llegara al binario, porque con él
// aparecerían go.yaml.in/yaml/v3 y santhosh-tekuri/jsonschema (research.md D2).
package skills
