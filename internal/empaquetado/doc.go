// Package empaquetado es la lógica del paso que empaqueta las dos piezas con
// las que kitlegal se instala sin terminal (docs/ADR/0035; H22 research.md D1):
// kitlegal.mcpb, la extensión de escritorio con el servidor MCP dentro, y
// kitlegal-plugin.zip, el plugin de Claude con las skills y, dentro, esa misma
// extensión. Con el plugin basta: la extensión se publica también suelta, para
// quien solo quiera las herramientas (ADR 0035, «Prueba con un solo plugin»).
// Su programa es cmd/empaquetar, que solo inyecta los argumentos y la salida de
// error y termina con el código que devuelve Ejecutar.
//
// Tiene dos órdenes (contracts/paso.md §1 de H22):
//
//   - piezas escribe los dos zips a partir de los dos binarios que recibe, que
//     copia sin mirarlos, del icono, de las herramientas que el registro de
//     applets anuncia por MCP y de las skills empotradas (piezas.go);
//   - catalogo escribe el catálogo de una versión, con el plugin de la release
//     de esa versión dentro, como carpeta (catalogo.go).
//
// Los textos con los que las piezas se presentan viven en textos.go, y en
// ningún otro sitio (H22 FR-015).
//
// El paso es reproducible: no pide nada a la red, no lee más que lo que se le
// nombra y no guarda estado, y dos ejecuciones con las mismas entradas dan los
// mismos bytes (H22 FR-004; research.md D8). No escribe nada en la salida
// estándar: sus órdenes escriben ficheros (research.md D16).
//
// Es un programa de construcción del repositorio, no un applet: lo ejecutan
// goreleaser y los flujos de la release con `go run`, compilado del mismo árbol
// que los binarios que empaqueta, y por eso compone con internal/app —de donde
// salen las herramientas— y con el paquete raíz —que lleva lo empotrado—. El
// binario distribuido no enlaza este paquete, y TestElBinarioNoEnlazaElPaso
// falla si algún día llegara a él (H22 FR-001; research.md D17).
package empaquetado
