// Package territorio resuelve un municipio de España, por su nombre o por su
// código INE, al terreno que lo rodea: su provincia, su comunidad autónoma, el
// DIR3 de su ayuntamiento, su régimen y los boletines que le aplican, con la
// cobertura de lo que el registro tiene configurado y verificado (FR-006,
// FR-020, data-model §2).
//
// El paquete no lee ficheros: recibe los de data/territorio/ ya leídos, como
// bytes (Fuentes), y quien los empaqueta es otro (research.md D3). Cargar los
// decodifica, comprueba su forma y su integridad entre ficheros y construye el
// Registro; Resolver no vuelve a leer nada.
//
// Lo que garantiza:
//
//   - Cargar exige los seis puntos de integridad de data-model §2.1 y todo lo
//     que el territorio resuelto va a emitir; unas fuentes con defectos no
//     cargan, y su error, que nombra cada fichero y cada dato, no lleva clase
//     de usuario: es un defecto de composición, no algo que quien pregunta
//     pueda corregir.
//   - Ningún municipio es un caso especial: las formas por las que se reconoce
//     un nombre se derivan de su nombre oficial con reglas generales, y lo que
//     cambia de un territorio a otro vive en su fichero de comunidad (FR-024,
//     FR-053).
//   - Un nombre que lleva a varios municipios es una ambigüedad declarada con
//     todos sus candidatos, nunca una elección (FR-013, FR-014).
//   - El territorio resuelto lleva siempre sus ocho claves, cada dato con su
//     source; nunca un DIR3 que no esté verificado ni un boletín que no esté
//     configurado, y su cobertura no tiene ningún valor que signifique «no
//     existe» (FR-008, FR-020 a FR-023).
//   - Una consulta que no se resuelve devuelve un error de clase «argumentos»
//     o «no encontrado», que nombra la entrada; ninguna otra (FR-010 a
//     FR-016).
//
// Es dominio puro: no hace entrada ni salida y no importa más paquetes
// internos que internal/core/ids, internal/core/schema e internal/core/grafo,
// por el vocabulario con el que declara lo que observa (regla R1 de
// docs/ROADMAP.md §2). Un Registro no cambia después de cargarse y puede
// usarse desde varias goroutines a la vez.
package territorio
