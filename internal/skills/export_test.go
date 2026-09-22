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
