package skills

// CompilarEsquemaDeNormas es compilarEsquemaDeNormas para los tests del paquete
// externo: con él fijan, sobre un directorio temporal, los errores con los que no
// se obtiene el esquema de data/normas.yaml, que la ruta constante del paquete no
// da nunca.
var CompilarEsquemaDeNormas = compilarEsquemaDeNormas

// CompilarEsquemaDelTerritorio es compilarEsquemaDelTerritorio para los tests
// del paquete externo, por la misma razón: las rutas constantes de los esquemas
// de data/territorio/ no dan nunca sus errores.
var CompilarEsquemaDelTerritorio = compilarEsquemaDelTerritorio

// CompilarEsquemaDeJerarquia es compilarEsquemaDeJerarquia para los tests del
// paquete externo, por la misma razón: la ruta constante del esquema de
// data/jerarquia.yaml no da nunca sus errores.
var CompilarEsquemaDeJerarquia = compilarEsquemaDeJerarquia

// EnlacesDeMarkdown, FicherosEnlazados y FaltaDelIndice son enlacesDeMarkdown,
// ficherosEnlazados y faltaDelIndice para los tests del paquete externo: con
// ellos fijan, sobre textos sueltos, lo que los gates de references/ leen de un
// fichero Markdown (contracts/gates-de-skills-check.md §2.3), cuyos casos límite
// no caben en las skills de un árbol de prueba sin que lo tape otro defecto.
var (
	EnlacesDeMarkdown = enlacesDeMarkdown
	FicherosEnlazados = ficherosEnlazados
	FaltaDelIndice    = faltaDelIndice
)

// EntradasQueNoSonFicherosMd es entradasQueNoSonFicherosMd para los tests del
// paquete externo: con ella TestReferenciasAMano fija, sobre el mismo árbol, que
// lo que no es una referencia a mano por no ser un fichero regular .md es lo que
// da el defecto de FR-006.
var EntradasQueNoSonFicherosMd = entradasQueNoSonFicherosMd
