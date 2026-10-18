// Package plazos calcula la fecha en que vence un plazo administrativo o
// procesal, con los festivos de los municipios que cuentan y la declaración de
// hasta dónde llegan los datos (FR-001 a FR-012, FR-020 a FR-026, data-model
// §1).
//
// El paquete no lee ficheros: recibe los de data/plazos.yaml y data/festivos/
// ya leídos, como bytes, y quien los empaqueta es otro (research D1). LeerTipos
// lee los siete tipos de plazo de FR-008, cada uno con su cómputo, su duración
// cuando la fija y los preceptos que la fundan. CargarCalendario lee los
// festivos nacionales y autonómicos de las resoluciones del BOE y los locales
// de los boletines de cada comunidad, y construye el Calendario, que responde
// sin volver a leer nada: si un día es inhábil para unos municipios, qué
// festivos lo hacen inhábil, qué festivos de una parte de la comunidad caen en
// él y qué cubren los datos en unos años. EscribirFestivosNacionales y
// EscribirFestivosLocales escriben los dos ficheros de data/festivos/ en su
// forma canónica, la que la derivación compara byte a byte con lo grabado
// (contracts/datos-de-plazos.md §1).
//
// Lo que garantiza:
//
//   - Los tipos son los siete de FR-008 y ninguno más, y cada uno lleva lo que
//     el resultado nombra: data/plazos.yaml no tiene otro lector ni esquema, así
//     que LeerTipos es su única definición (research D17).
//   - Un festivo de una parte de la comunidad no hace inhábil el día en toda
//     ella: el calendario lo da aparte, para que se avise (FR-011, FR-025).
//   - La cobertura nunca da por inexistentes los festivos que no tiene: dice,
//     por ámbito y por tramos de años, hasta qué disposición llegan los datos o
//     por qué faltan, de un vocabulario cerrado (FR-022, FR-026).
//   - Un fichero de data/festivos/ o data/plazos.yaml que no se puede leer es
//     un error que nombra el fichero y no lleva clase de usuario: el kernel lo da
//     como inesperado, código 1. Los de data/festivos/ no tienen reglas de carga
//     propias; lo que llevan lo garantiza la derivación (research D25).
//   - El Calendario no cambia después de cargarse y puede usarse desde varias
//     goroutines a la vez.
package plazos
