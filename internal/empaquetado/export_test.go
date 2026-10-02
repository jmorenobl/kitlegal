package empaquetado

// PiezasAEscribir es piezasAEscribir para los tests del paquete externo: con él
// dan al paso unas herramientas y unas skills que no son las de producción, y
// una descripción corta y un icono que no son los del repositorio, que la
// composición de Ejecutar no da nunca.
type PiezasAEscribir = piezasAEscribir

// EscribirPiezas es escribirPiezas para los tests del paquete externo: lo que
// escribe las dos piezas con lo que se le da.
var EscribirPiezas = escribirPiezas

// DocumentoDelCatalogo es documentoDelCatalogo para los tests del paquete
// externo: con él leen el catálogo de una versión y una huella, y fijan el
// error de una versión vacía, que las banderas de Ejecutar no dejan pasar.
var DocumentoDelCatalogo = documentoDelCatalogo
