// Command kitlegal-e2e es el binario contra el que se ejecuta el test de
// extremo a extremo: el **kernel real** —el mismo internal/app que enlaza el
// binario que se publica— más el registro de los applets de ejemplo. Lo único
// que cambia entre este binario y el distribuido es qué applets se registran
// (FR-009, contracts/registro-y-describe.md §3).
//
// Es la segunda —y última— raíz de composición del proyecto, y por eso es uno de
// los dos únicos sitios del árbol donde se nombran os.Exit, os.Stdout y
// os.Stderr: no hay forma de que un `package main` propague un código de salida
// sin lo primero —retornar de main sale siempre con 0— ni de que inyecte los
// descriptores sin nombrarlos. La excepción del lint se acota a este directorio
// y no alcanza al paquete hermano de applets de ejemplo (research.md D18, D19).
//
// Los comodines de Go no descienden a un directorio testdata, así que este
// binario no se enlaza jamás en el artefacto distribuido, ni por descuido.
package main

import (
	"os"

	"github.com/jmorenobl/kitlegal/internal/app"
	"github.com/jmorenobl/kitlegal/internal/app/testdata/ejemplo"
)

// Datos de construcción. El e2e compila este paquete sin -ldflags, así que los
// valores que se ven son estos; el contrato de `version` que ejerce el binario
// distribuido lo comprueban los tests de cmd/kitlegal (D16).
var (
	version = "dev"
	commit  = "none"
	fecha   = "unknown"
)

// main es, línea por línea, el mismo que el del binario distribuido salvo el
// registro que inyecta: app.Main devuelve el código de salida y nunca termina el
// proceso, y acabar con él es lo único que este punto de entrada hace
// (contracts/reglas-de-arquitectura.md §2, R4).
func main() {
	os.Exit(app.Main(
		os.Args, registroDeEjemplo(), os.Stdout, os.Stderr, version, commit, fecha,
	))
}

// registroDeEjemplo construye el registro de este binario. Un registro inválido
// es un defecto de quien escribió el applet y **revienta al arrancar**, antes de
// atender ninguna invocación: no se convierte en un código de salida de usuario
// y no llega a la traducción única, que solo traduce fallos de quien invoca
// (FR-008, research.md D17).
func registroDeEjemplo() *app.Registro {
	registro, err := ejemplo.Registro()
	if err != nil {
		panic(err)
	}

	return registro
}
