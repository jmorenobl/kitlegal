package httpx

// version es la versión con la que el paquete se identifica en la red. El valor
// por omisión es el mismo que declaran cmd/kitlegal/main.go y el main del
// binario de extremo a extremo, de modo que una construcción sin datos
// inyectados —go run, go test— se identifique con la versión que ese mismo
// binario imprime en su verbo version (FR-007).
//
// El Makefile la inyecta con -X en la misma orden que inyecta main.version, y
// con la misma VERSION, así que las dos llegan siempre juntas. Es el único
// estado del paquete que vive en una variable, y solo se lee: no es exportada
// —nada de fuera puede vaciarla (FR-008)— ni la escribe ninguna función.
var version = "dev"

// AgenteDeUsuario devuelve la identificación del proyecto con la forma exacta
// que exige FR-006: el nombre, la versión de este binario y la dirección donde
// quien administra un sitio puede averiguar qué es lo que le está pidiendo
// datos. No se construye desde fuera, no se configura y no admite valor vacío,
// porque identificarse no es una opción del cliente sino parte de la forma de
// pedir (FR-008, constitución §I).
//
// La versión no está escrita a mano aquí: sale de version, que es lo que impide
// que la identificación envejezca sola cuando el binario cambie de versión
// (FR-007, research.md D5).
func AgenteDeUsuario() string {
	return "kitlegal/" + version + " (+https://kitlegal.es/bot)"
}
